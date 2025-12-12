// Himiko Discord Bot
// Copyright (C) 2025 Himiko Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

// Track represents a music track
type Track struct {
	Title     string
	URL       string
	Duration  int
	Thumbnail string
	Requester string
	IsLocal   bool
}

// MusicPlayer handles audio playback for a guild
type MusicPlayer struct {
	guildID    string
	voiceConn  *discordgo.VoiceConnection
	encoding   *dca.EncodeSession
	streaming  *dca.StreamingSession
	queue      []*Track
	nowPlaying *Track
	volume     int
	mu         sync.RWMutex
	stopChan   chan bool
	isPlaying  bool
	isPaused   bool
}

// MusicManager manages music players across guilds
type MusicManager struct {
	players map[string]*MusicPlayer
	mu      sync.RWMutex
}

// NewMusicManager creates a new music manager
func NewMusicManager() *MusicManager {
	return &MusicManager{
		players: make(map[string]*MusicPlayer),
	}
}

// GetPlayer gets or creates a player for a guild
func (m *MusicManager) GetPlayer(guildID string) *MusicPlayer {
	m.mu.Lock()
	defer m.mu.Unlock()

	if player, exists := m.players[guildID]; exists {
		return player
	}

	player := &MusicPlayer{
		guildID:   guildID,
		queue:     make([]*Track, 0),
		volume:    50,
		stopChan:  make(chan bool, 1),
		isPlaying: false,
		isPaused:  false,
	}
	m.players[guildID] = player
	return player
}

// RemovePlayer removes a player for a guild
func (m *MusicManager) RemovePlayer(guildID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if player, exists := m.players[guildID]; exists {
		player.Stop()
		player.Disconnect()
		delete(m.players, guildID)
	}
}

// VideoInfo holds info extracted from yt-dlp
type VideoInfo struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Duration  int    `json:"duration"`
	Thumbnail string `json:"thumbnail"`
}

// ExtractInfo extracts video info using yt-dlp
func ExtractInfo(url string) (*VideoInfo, error) {
	// Check if it's a local file
	if isLocalFile(url) {
		return extractLocalFileInfo(url)
	}

	args := []string{
		"--dump-json",
		"--no-playlist",
		"--format", "bestaudio",
	}

	// Add API keys if available
	if youtubeKey := os.Getenv("YOUTUBE_API_KEY"); youtubeKey != "" {
		args = append(args, "--username", "oauth2", "--password", "")
	}

	if soundcloudAuth := os.Getenv("SOUNDCLOUD_AUTH_TOKEN"); soundcloudAuth != "" {
		args = append(args, "--add-header", "Authorization:OAuth "+soundcloudAuth)
	}

	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to extract info: %w", err)
	}

	var info VideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse video info: %w", err)
	}

	return &info, nil
}

func isLocalFile(path string) bool {
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func extractLocalFileInfo(path string) (*VideoInfo, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	fileName := filepath.Base(path)
	title := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	return &VideoInfo{
		Title:     title,
		URL:       path,
		Duration:  0,
		Thumbnail: "",
	}, nil
}

// Connect joins a voice channel
func (p *MusicPlayer) Connect(s *discordgo.Session, channelID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.voiceConn != nil {
		return nil
	}

	vc, err := s.ChannelVoiceJoin(p.guildID, channelID, false, true)
	if err != nil {
		return fmt.Errorf("failed to join voice channel: %w", err)
	}

	p.voiceConn = vc
	return nil
}

// Disconnect leaves the voice channel
func (p *MusicPlayer) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.voiceConn != nil {
		if err := p.voiceConn.Disconnect(); err != nil {
			return err
		}
		p.voiceConn = nil
	}

	p.stopInternal()
	return nil
}

// AddTrack adds a track to the queue
func (p *MusicPlayer) AddTrack(track *Track) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue = append(p.queue, track)
}

// GetQueue returns a copy of the queue
func (p *MusicPlayer) GetQueue() []*Track {
	p.mu.RLock()
	defer p.mu.RUnlock()

	queueCopy := make([]*Track, len(p.queue))
	copy(queueCopy, p.queue)
	return queueCopy
}

// RemoveTrack removes a track at the given position
func (p *MusicPlayer) RemoveTrack(position int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if position < 0 || position >= len(p.queue) {
		return errors.New("invalid position")
	}

	p.queue = append(p.queue[:position], p.queue[position+1:]...)
	return nil
}

// MoveToTop moves a track to the top of the queue
func (p *MusicPlayer) MoveToTop(position int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if position < 0 || position >= len(p.queue) {
		return errors.New("invalid position")
	}

	track := p.queue[position]
	p.queue = append(p.queue[:position], p.queue[position+1:]...)
	p.queue = append([]*Track{track}, p.queue...)

	return nil
}

// ClearQueue clears all tracks from the queue
func (p *MusicPlayer) ClearQueue() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue = make([]*Track, 0)
}

// Play starts playback
func (p *MusicPlayer) Play() error {
	p.mu.Lock()
	if p.isPlaying {
		p.mu.Unlock()
		return nil
	}

	if len(p.queue) == 0 {
		p.mu.Unlock()
		return errors.New("queue is empty")
	}

	if p.voiceConn == nil {
		p.mu.Unlock()
		return errors.New("not connected to voice channel")
	}

	p.isPlaying = true
	p.mu.Unlock()

	go p.playLoop()
	return nil
}

func (p *MusicPlayer) playLoop() {
	for {
		p.mu.Lock()
		if len(p.queue) == 0 {
			p.isPlaying = false
			p.nowPlaying = nil
			p.mu.Unlock()
			return
		}

		track := p.queue[0]
		p.queue = p.queue[1:]
		p.nowPlaying = track
		p.mu.Unlock()

		if err := p.playTrack(track); err != nil {
			fmt.Printf("Error playing track: %v\n", err)
		}

		select {
		case <-p.stopChan:
			p.mu.Lock()
			p.isPlaying = false
			p.nowPlaying = nil
			p.mu.Unlock()
			return
		default:
		}
	}
}

func (p *MusicPlayer) playTrack(track *Track) error {
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 128
	options.Application = "audio"
	options.Volume = p.volume

	// Handle local files
	if track.IsLocal {
		encodeSession, err := dca.EncodeFile(track.URL, options)
		if err != nil {
			return fmt.Errorf("failed to encode local file: %w", err)
		}
		defer encodeSession.Cleanup()

		p.mu.Lock()
		p.encoding = encodeSession
		done := make(chan error)
		streamSession := dca.NewStream(encodeSession, p.voiceConn, done)
		p.streaming = streamSession
		p.mu.Unlock()

		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				return fmt.Errorf("streaming error: %w", err)
			}
		case <-p.stopChan:
			p.mu.Lock()
			if p.streaming != nil {
				p.streaming.SetPaused(true)
			}
			if p.encoding != nil {
				p.encoding.Cleanup()
			}
			p.mu.Unlock()
			return nil
		}

		return nil
	}

	// Handle online URLs with yt-dlp
	args := []string{
		"--format", "bestaudio",
		"--output", "-",
		"--no-playlist",
	}

	if youtubeKey := os.Getenv("YOUTUBE_API_KEY"); youtubeKey != "" {
		args = append(args, "--username", "oauth2", "--password", "")
	}

	if soundcloudAuth := os.Getenv("SOUNDCLOUD_AUTH_TOKEN"); soundcloudAuth != "" {
		args = append(args, "--add-header", "Authorization:OAuth "+soundcloudAuth)
	}

	args = append(args, track.URL)

	cmd := exec.Command("yt-dlp", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	encodeSession, err := dca.EncodeMem(stdout, options)
	if err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to encode audio: %w", err)
	}
	defer encodeSession.Cleanup()

	p.mu.Lock()
	p.encoding = encodeSession
	done := make(chan error)
	streamSession := dca.NewStream(encodeSession, p.voiceConn, done)
	p.streaming = streamSession
	p.mu.Unlock()

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			return fmt.Errorf("streaming error: %w", err)
		}
	case <-p.stopChan:
		p.mu.Lock()
		if p.streaming != nil {
			p.streaming.SetPaused(true)
		}
		if p.encoding != nil {
			p.encoding.Cleanup()
		}
		p.mu.Unlock()
		return nil
	}

	cmd.Wait()
	return nil
}

// Skip skips the current track
func (p *MusicPlayer) Skip() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isPlaying {
		return errors.New("nothing is playing")
	}

	if p.streaming != nil {
		p.streaming.SetPaused(true)
	}

	if p.encoding != nil {
		p.encoding.Cleanup()
	}

	return nil
}

// Stop stops playback completely
func (p *MusicPlayer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopInternal()
}

func (p *MusicPlayer) stopInternal() {
	if p.isPlaying {
		select {
		case p.stopChan <- true:
		default:
		}
	}

	if p.streaming != nil {
		p.streaming.SetPaused(true)
	}

	if p.encoding != nil {
		p.encoding.Cleanup()
	}

	p.isPlaying = false
	p.nowPlaying = nil
}

// Pause pauses playback
func (p *MusicPlayer) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isPlaying {
		return errors.New("nothing is playing")
	}

	if p.streaming != nil {
		p.streaming.SetPaused(true)
		p.isPaused = true
	}

	return nil
}

// Resume resumes playback
func (p *MusicPlayer) Resume() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isPaused {
		return errors.New("player is not paused")
	}

	if p.streaming != nil {
		p.streaming.SetPaused(false)
		p.isPaused = false
	}

	return nil
}

// SetVolume sets the volume (0-100)
func (p *MusicPlayer) SetVolume(volume int) error {
	if volume < 0 || volume > 100 {
		return errors.New("volume must be between 0 and 100")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.volume = volume
	return nil
}

// NowPlaying returns the currently playing track
func (p *MusicPlayer) NowPlaying() *Track {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.nowPlaying
}

// IsPlaying returns whether the player is playing
func (p *MusicPlayer) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.isPlaying
}

// IsPaused returns whether the player is paused
func (p *MusicPlayer) IsPaused() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.isPaused
}

// IsConnected returns whether the player is connected to voice
func (p *MusicPlayer) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.voiceConn != nil
}

// MusicPermLevel represents permission levels for music commands
type MusicPermLevel int

const (
	MusicPermUser MusicPermLevel = iota
	MusicPermDJ
	MusicPermMod
	MusicPermAdmin
)

// GetMusicPermLevel gets a user's music permission level
func GetMusicPermLevel(s *discordgo.Session, guildID, userID string, djRoleID, modRoleID *string) MusicPermLevel {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return MusicPermUser
	}

	// Check for admin permissions
	perms, err := s.UserChannelPermissions(userID, guildID)
	if err == nil && (perms&discordgo.PermissionAdministrator != 0 || perms&discordgo.PermissionManageServer != 0) {
		return MusicPermAdmin
	}

	// Check for mod role
	if modRoleID != nil && *modRoleID != "" {
		for _, roleID := range member.Roles {
			if roleID == *modRoleID {
				return MusicPermMod
			}
		}
	}

	// Check for DJ role
	if djRoleID != nil && *djRoleID != "" {
		for _, roleID := range member.Roles {
			if roleID == *djRoleID {
				return MusicPermDJ
			}
		}
	}

	return MusicPermUser
}

// GetUserVoiceChannel gets the voice channel a user is in
func GetUserVoiceChannel(s *discordgo.Session, guildID, userID string) (string, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return "", err
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID, nil
		}
	}

	return "", fmt.Errorf("user not in voice channel")
}
