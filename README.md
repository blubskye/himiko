<p align="center">
  <img src="himiko.png" alt="Himiko" width="400"/>
</p>

<h1 align="center">💉 Himiko Discord Bot 💉</h1>

<p align="center">
  <em>"I just wanna love you, wanna be loved~"</em>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go"/>
  <img src="https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite"/>
  <img src="https://img.shields.io/badge/Discord-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Discord"/>
  <img src="https://img.shields.io/badge/License-AGPL--3.0-red?style=for-the-badge" alt="License"/>
</p>

---

## 🩸 About Himiko

A feature-rich Discord bot written in Go with SQLite storage, named after everyone's favorite blood-obsessed villain! She's cute, she's crazy, and she'll manage your server with deadly efficiency~ 💕

> *"Let me help you... I promise I won't bite~ Much."*

---

## ✨ Features

### 🔪 Administration
- **Moderation:** Kick, ban, unban, softban, hackban
- **Timeout:** Timeout and remove timeout
- **Messages:** Purge messages (with filters)
- **Channel Control:** Slowmode, lock/unlock channels
- **Warning System:** Track troublemakers~
- **View Bans:** See who's been naughty

### 🎀 XP & Leveling System
- **Track Activity:** Users earn XP by chatting
- **Leaderboards:** See who's the most active!
- **Level Roles:** Auto-assign roles at level milestones
- **Voice XP:** Earn XP in voice channels too~
- **Admin Controls:** Set levels, add XP, mass XP operations

### 🛡️ Auto-Moderation
- **Regex Filters:** Custom pattern matching with actions (delete/warn/ban)
- **Test Filters:** Test patterns before enabling
- **Per-Channel Config:** Disable logging for specific channels
- **Spam Filter:** Limit mentions, links, and emojis with configurable actions

### 🚨 Anti-Raid Protection
- **Raid Detection:** Automatic detection of mass joins
- **Auto-Silence Modes:** Log, alert, raid-only, or all-joins silencing
- **Server Lockdown:** Automatic verification level raise during raids
- **Silence/Unsilence:** Manual and timed user silencing
- **Ban Raid:** Bulk ban detected raid users
- **Account Age Alerts:** Flag new accounts joining

### 🔥 Advanced Anti-Spam (Pressure System)
- **Pressure-Based Detection:** Accumulates spam pressure per user
- **Configurable Penalties:** Images, links, pings, length, repeats
- **Decay System:** Pressure naturally decreases over time
- **Actions:** Delete, warn, silence, kick, or ban spammers

### 📊 Moderation Stats
- **Track Mod Actions:** Import and track bans, kicks, timeouts
- **Mod Stats:** See which moderators are most active
- **User History:** View moderation history for specific users

### 🧹 Auto-Clean System
- **Channel Cleaning:** Automatically clean channels on schedule
- **Warning Messages:** Warn users before cleaning
- **Preserve Options:** Keep images if desired

### 📝 Logging System
- **Message Logs:** Deleted/edited messages
- **Voice Logs:** Join/leave events
- **User Changes:** Nicknames, avatars
- **Configurable:** Enable/disable each log type

### 🎲 Fun Commands
- 8-ball, dice rolls, coinflip
- Rock Paper Scissors
- Random number generator
- Jokes, rate things, ship compatibility
- IQ/gay/PP tests (joke commands)
- Social interactions (hug, slap, pat, kiss)
- Would you rather, truth or dare
- Choose between options

### 📝 Text Transformations
- ASCII art, Zalgo text
- Reverse, upside down
- Morse code, Vaporwave
- OwO, mock text, Leet speak (1337)
- Regional indicators (emoji letters)
- Spoiler each character
- Encode/decode (base64, hex, binary, rot13)
- Codeblock wrapper, Hyperlink creator

### 🖼️ Images
- Random animal images (cat, dog, fox, bird, bunny, duck, koala, panda)
- User avatar and banner
- Server icon
- Cat and dog facts
- Random memes from Reddit

### 🔧 Utility
- Ping (latency check)
- Snipe deleted messages
- AFK status, Reminders
- Scheduled messages, Polls
- Custom embeds
- Clean your messages
- First message in channel
- Bot uptime, Say command
- Steal emoji, Simple math

### ℹ️ Information
- User/Server/Channel/Role info
- Emoji info, Bot info
- Invite info, Role list
- Member count

### 🔍 Lookup
- Weather, Urban Dictionary
- Wikipedia, IP address lookup
- Cryptocurrency prices
- Minecraft server status
- GitHub users, NPM packages
- Color information

### 🎰 Random
- Advice, quotes, facts
- Trivia questions
- Would you rather
- Truth or dare
- Never have I ever
- Dad jokes
- Password generator

### 🛠️ Tools
- URL shortener (TinyURL)
- QR code generator
- Discord timestamp generator
- Character counter
- Snowflake decoder
- Server list (bot owner)
- Permission viewer
- Raw message content
- Message link generator
- **Ban Export/Import** - Share ban lists between servers!

### 🎵 Music System
- **URL Playback:** Play from YouTube, SoundCloud, and more via yt-dlp
- **Local Library:** Play files from configured music folders
- **Queue Management:** Add, remove, move tracks in queue
- **Playback Controls:** Play, pause, resume, skip, stop
- **Volume Control:** Adjust playback volume (0-100)
- **DJ/Mod Roles:** Permission system for music commands
- **History:** Track recently played songs
- **Search:** Search local music library

### 🎫 Ticket System
- **Submit Tickets:** Users can report issues to staff
- **Configurable Channel:** Set where tickets are forwarded
- **Clean Interface:** User messages are ephemeral, staff sees formatted embed

### 💬 Mention Responses
- **Custom Triggers:** Set responses when bot is mentioned with keywords
- **Image Support:** Include images in responses

### 📨 Join DM Messages
- **Welcome DMs:** Send customizable DMs to new members
- **Embed Support:** Include title and message with placeholders

### ⚙️ Settings
- Custom prefix
- Mod log channel
- Welcome messages
- View server settings

### 🤖 AI Integration
- Ask AI questions (requires OpenAI API key or compatible endpoint)

### 🔄 Auto-Update System
- **Update Checking:** Automatically checks for new versions on startup
- **Periodic Checks:** Configurable interval for background update checks (default: 24 hours)
- **Channel Notifications:** Optionally post update notifications to a channel
- **Manual Updates:** Use `/update check` to check for updates
- **Easy Apply:** Use `/update apply` to download and install updates
- **Auto-Relaunch:** Bot automatically restarts with new version after update
- **Config Preservation:** Updates keep your config.json intact (max 3 backups)
- **Configurable:** Enable/disable auto-update, auto-apply, check interval, and notification channel

### 🌐 Web Dashboard
- **Server Management:** Visual dashboard to manage servers and settings
- **Stats Overview:** View bot statistics, server counts, member counts
- **Guild Settings:** Configure prefix, welcome messages, and more per-server
- **Toggle Control:** Enable/disable via `/webserver on` and `/webserver off`
- **NGINX Ready:** Designed to work behind NGINX reverse proxy
- **Local by Default:** Binds to 127.0.0.1 for security, configurable for remote access

### 📈 Live Stats & Monitoring
- **Real-Time Dashboard:** View live metrics with auto-updating values
- **Memory Monitoring:** Track Alloc, Sys memory, and GC runs
- **Discord Stats:** Guilds, members, channels, heartbeat latency
- **Activity Metrics:** Commands processed, messages seen, rates per minute
- **Interactive Charts:** Memory and activity graphs (last hour)
- **Database Stats:** File size and table row counts
- **SSE Updates:** Server-Sent Events for instant 5-second updates
- **Uptime Counter:** Live uptime display
- **Discord Command:** `botstats` for quick stats in Discord (Owner only)

### 🐛 Debug Mode
- **Full Stack Traces:** Enable verbose logging with complete stack traces
- **Error Details:** Get detailed error information for troubleshooting
- **Memory Stats:** View memory and goroutine statistics
- **Caller Info:** Track exactly where errors originate

### 🚫 Bot Management (Owner Only)
- Bot-level bans for users/servers
- DM forwarding to designated channels

---

## 💉 Setup

*"Let me help you get started~"*

### 1. Clone the repository
```bash
git clone https://github.com/blubskye/himiko.git
cd himiko
```

### 2. Configure
Copy `config.example.json` to `config.json` and fill in your details:

```json
{
  "token": "YOUR_BOT_TOKEN_HERE",
  "prefix": "/",
  "database_path": "himiko.db",
  "owner_id": "YOUR_DISCORD_USER_ID",
  "owner_ids": ["YOUR_DISCORD_USER_ID", "OPTIONAL_SECOND_OWNER_ID"],
  "apis": {
    "weather_api_key": "",
    "google_api_key": "",
    "spotify_client_id": "",
    "spotify_client_secret": "",
    "openai_api_key": "",
    "openai_base_url": "https://api.openai.com/v1",
    "openai_model": "gpt-3.5-turbo",
    "youtube_api_key": "",
    "soundcloud_auth_token": ""
  },
  "features": {
    "dm_logging": false,
    "command_history": true,
    "delete_timer": 0,
    "webhook_notify": false,
    "webhook_url": "",
    "auto_update": true,
    "auto_update_apply": false,
    "update_check_hours": 24,
    "update_notify_channel": "",
    "debug_mode": false
  },
  "webserver": {
    "enabled": false,
    "port": 8080,
    "host": "127.0.0.1",
    "secret_key": "",
    "allow_remote": false
  }
}
```

### 3. Build and run
```bash
go build ./cmd/himiko
./himiko
```

---

## 🔄 Running as a Service (Auto-start on Boot)

*"I'll always be here when you wake up... waiting for you~"* 💉

The `scripts/` directory contains helper scripts to run Himiko in the background with the ability to check on her output.

### 💻 Quick Start (Manual tmux - Linux/macOS/BSD)

```bash
# Wake Himiko up in a tmux session~
./scripts/himiko-tmux.sh start

# Connect to Himiko's terminal
./scripts/himiko-tmux.sh attach

# Check if Himiko is running
./scripts/himiko-tmux.sh status

# Let Himiko rest...
./scripts/himiko-tmux.sh stop
```

> 💡 To detach from tmux without stopping Himiko: Press `Ctrl+B`, then `D`

### 🐧 Linux (systemd)

*"I'll start automatically... because I just wanna be with you~"*

1. Edit the service file to match your setup:
   ```bash
   nano scripts/himiko-bot.service
   ```
   Change `YOUR_USER` to your username and `/path/to/himiko` to the actual path.

2. Install and enable:
   ```bash
   sudo cp scripts/himiko-bot.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable himiko-bot
   sudo systemctl start himiko-bot
   ```

3. Attach to Himiko's terminal:
   ```bash
   tmux attach -t himiko-bot
   ```

4. Check on Himiko:
   ```bash
   sudo systemctl status himiko-bot
   ```

### 😈 FreeBSD (rc.d)

*"Even on BSD... I'll find a way to be with you~"*

1. Install the rc script:
   ```bash
   sudo cp scripts/himiko-bot-freebsd /usr/local/etc/rc.d/himiko_bot
   sudo chmod +x /usr/local/etc/rc.d/himiko_bot
   ```

2. Configure in `/etc/rc.conf`:
   ```bash
   sudo sysrc himiko_bot_enable=YES
   sudo sysrc himiko_bot_user="YOUR_USER"
   sudo sysrc himiko_bot_dir="/path/to/himiko"
   ```

3. Start Himiko:
   ```bash
   sudo service himiko_bot start
   ```

4. Attach to Himiko's terminal:
   ```bash
   su - YOUR_USER -c "tmux attach -t himiko-bot"
   ```

### 🪟 Windows

*"I'll run on Windows too... nothing can keep us apart~"*

#### Quick Start (Visible Console)
```batch
scripts\himiko-start.bat
```

#### PowerShell Service Manager
```powershell
# Start Himiko in the background
.\scripts\himiko-service.ps1 start

# Check status
.\scripts\himiko-service.ps1 status

# View output (Ctrl+C to detach)
.\scripts\himiko-service.ps1 attach

# Stop Himiko
.\scripts\himiko-service.ps1 stop

# Add to Windows startup (runs on login)
.\scripts\himiko-service.ps1 install

# Remove from startup
.\scripts\himiko-service.ps1 uninstall
```

#### Hidden Startup (No Console Window)
Double-click `scripts\himiko-hidden.vbs` to start Himiko without a console window.
To add to startup manually:
1. Press `Win+R`, type `shell:startup`, press Enter
2. Create a shortcut to `himiko-hidden.vbs` in that folder

---

## 🎀 Getting a Bot Token

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application (name it Himiko, of course~)
3. Go to the "Bot" section
4. Click "Add Bot"
5. Copy the token
6. Enable **all** Privileged Gateway Intents:
   - Presence Intent
   - Server Members Intent
   - Message Content Intent

---

## 💕 Inviting Himiko

Use the `/invite` command once she's running, or construct the URL:

```
https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=8&scope=bot%20applications.commands
```

---

## 📋 Requirements

- Go 1.21+
- GCC (for SQLite compilation with go-sqlite3)
- FFmpeg (for music playback)
- yt-dlp (for YouTube/SoundCloud support)

---

## 🩸 Commands List

| Category | Commands |
|----------|----------|
| **Admin** | kick, ban, unban, softban, hackban, timeout, untimeout, purge, slowmode, lock, unlock, warn, warnings, clearwarnings, bans |
| **XP** | xp, rank, leaderboard, setlevel, setxp, addxp, massaddxp |
| **Ranks** | addrank, removerank, listranks, syncranks, applyranks |
| **Voice XP** | voicexp (enable/disable/rate/interval/ignoreafk/status) |
| **Filters** | addfilter, removefilter, listfilters, testfilter |
| **AutoClean** | autoclean (add/remove/list), setcleanmessage, setcleanimage |
| **Logging** | setlogchannel, togglelogging, logconfig, disablechannellog, enablechannellog, logstatus |
| **Fun** | 8ball, dice, coinflip, rps, random, joke, rate, ship, iq, gayrate, pp, hug, slap, pat, kiss, wyr, tod, choose |
| **Text** | ascii, zalgo, reverse, upsidedown, morse, vaporwave, owo, mock, leet, regional, spoilertext, encode, decode, codeblock, hyperlink |
| **Images** | cat, dog, fox, bird, bunny, duck, koala, panda, avatar, banner, servericon, catfact, dogfact, meme |
| **Utility** | ping, snipe, afk, remind, schedule, poll, embed, clean, firstmessage, uptime, say, stealemoji, math |
| **Info** | userinfo, serverinfo, channelinfo, roleinfo, emojiinfo, botinfo, stats, inviteinfo, rolelist, membercount |
| **Lookup** | weather, urban, wiki, ip, crypto, minecraft, github, npm, color |
| **Random** | advice, quote, fact, trivia, wyr, tod, nhie, dadjoke, password |
| **Tools** | tinyurl, qrcode, timestamp, charcount, snowflake, servers, permissions, raw, messagelink |
| **BanExport** | exportbans, importbans, scanbans |
| **ModStats** | modstats, importmodhistory, modhistory |
| **SpamFilter** | spamfilter (status/enable/disable/set) |
| **Anti-Raid** | antiraid (status/enable/disable/set/setrole/setalert/autosilence), silence, unsilence, getraid, banraid, lockdown |
| **Anti-Spam** | antispam (status/enable/disable/set/penalties/setrole) |
| **Mentions** | mention (add/remove/list) |
| **Ticket** | ticket, setticket, disableticket, ticketstatus |
| **Settings** | setprefix, setmodlog, setwelcome, disablewelcome, setjoindm, disablejoindm, settings |
| **DM** | setdmchannel, disabledm, dmstatus |
| **BotBan** | botban, botunban, botbanlist |
| **AI** | ask |
| **Music** | play, skip, stop, pause, resume, queue, nowplaying, remove, clear, movetop, volume, join, leave, musicrole, folders, files, local, search, musicfolder, musichistory |
| **Update** | update (check/apply/version) |
| **WebServer** | webserver (on/off/status/config), botstats |
| **Misc** | help, command, tag, notify, history, about, invite, source |

---

## 💉 Source Code

This bot is licensed under **AGPL-3.0**. You can view the source code using the `/source` command or visiting:

**https://github.com/blubskye/himiko**

*"I'll always be transparent with you~ That's true love, right?"*

---

## 🩸 License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)** 💉

### 💉 What This Means For You~

*"I want to share everything with you... let's be open with each other~"*

The AGPL-3.0 is a **copyleft license** that ensures this software remains free and open. Here's what you need to know:

#### ✅ You CAN:
- 💉 **Use** this bot for any purpose (personal, commercial, whatever~)
- 🔪 **Modify** the code to your heart's content
- 📤 **Distribute** copies to others
- 🌐 **Run** it as a network service (like a public Discord bot)

#### 📋 You MUST:
- 📖 **Keep it open source** - If you modify and distribute this code, your version must also be AGPL-3.0
- 🔗 **Provide source access** - Users of your modified bot must be able to get the source code
- 📝 **State changes** - Document what you've modified from the original
- 💌 **Include license** - Keep the LICENSE file and copyright notices intact

#### 🌐 The Network Clause (This is the important part!):
*"Even if we're apart... I'll always find you~"*

Unlike regular GPL, **AGPL has a network provision**. This means:
- If you run a **modified version** of this bot as a public service (like hosting it for others to use on Discord)
- You **MUST** make your complete source code available to users
- The `/source` command in this bot helps satisfy this requirement!

#### ❌ You CANNOT:
- 🚫 Make it closed source
- 🚫 Remove the license or copyright notices
- 🚫 Use a different license for modified versions
- 🚫 Hide your modifications if you run it as a public service

#### 💡 In Simple Terms:
> *"If you use my code to create something, you must share it with everyone too~ That's true love, right?"*

This ensures that improvements to the bot benefit the entire community, not just one person.

```
Himiko Discord Bot
Copyright (C) 2025 Himiko Contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.
```

See the [LICENSE](LICENSE) file for the full legal text.

**Source Code:** https://github.com/blubskye/himiko

---

<p align="center">
  <em>Made with 💉 and obsessive love</em>
</p>

<p align="center">
  <img src="himiko.png" alt="Himiko" width="100"/>
</p>
