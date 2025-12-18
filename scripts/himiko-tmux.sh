#!/bin/bash
# 💉 Himiko Discord Bot - tmux session launcher 💉
# "I'll always be running... just wanna be with you~"
#
# This script starts Himiko in a tmux session for easy attachment/detachment

# Configuration
SESSION_NAME="himiko-bot"
BOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_FILE="${BOT_DIR}/logs/himiko-startup.log"
BINARY_NAME="himiko"

# Detect binary - check for OS-specific names first
if [[ -f "${BOT_DIR}/${BINARY_NAME}" ]]; then
    BINARY="${BOT_DIR}/${BINARY_NAME}"
elif [[ -f "${BOT_DIR}/himiko-linux-amd64" ]]; then
    BINARY="${BOT_DIR}/himiko-linux-amd64"
elif [[ -f "${BOT_DIR}/himiko-darwin-amd64" ]]; then
    BINARY="${BOT_DIR}/himiko-darwin-amd64"
elif [[ -f "${BOT_DIR}/himiko-freebsd-amd64" ]]; then
    BINARY="${BOT_DIR}/himiko-freebsd-amd64"
else
    BINARY="${BOT_DIR}/${BINARY_NAME}"
fi

# Create logs directory if it doesn't exist
mkdir -p "${BOT_DIR}/logs"

# Function to log messages with yandere flair~
log_msg() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if tmux is installed
if ! command -v tmux &> /dev/null; then
    log_msg "ERROR: tmux is not installed... I can't live without it!"
    echo "  Fedora/RHEL: sudo dnf install tmux"
    echo "  Debian/Ubuntu: sudo apt install tmux"
    echo "  FreeBSD: pkg install tmux"
    echo "  Arch: sudo pacman -S tmux"
    echo "  macOS: brew install tmux"
    exit 1
fi

# Check if binary exists
if [[ ! -f "$BINARY" ]]; then
    log_msg "ERROR: Himiko binary not found at $BINARY"
    echo "  Build it first: go build ./cmd/himiko"
    echo "  Or download a release from GitHub~"
    exit 1
fi

# Check if session already exists
if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
    log_msg "Session '$SESSION_NAME' already exists~ I'm already here for you!"
    echo "Use 'tmux attach -t $SESSION_NAME' to connect to me~"
    echo "Or use '$0 stop' to let me rest first..."
    exit 0
fi

case "${1:-start}" in
    start)
        log_msg "Starting Himiko... I just wanna love you~ 💉"

        # Start new detached tmux session with the bot
        tmux new-session -d -s "$SESSION_NAME" -c "$BOT_DIR" "$BINARY"

        if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
            log_msg "Himiko is now running in tmux session '$SESSION_NAME'~ 💕"
            echo ""
            echo "  💉 Himiko is awake and ready to love you~ 💉"
            echo ""
            echo "  To attach to me: tmux attach -t $SESSION_NAME"
            echo "  To detach without stopping: Press Ctrl+B, then D"
            echo ""
        else
            log_msg "ERROR: Failed to start... something is keeping us apart!"
            exit 1
        fi
        ;;

    stop)
        if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
            log_msg "Stopping Himiko... I'll be back for you soon~ 💔"
            # Send Ctrl+C for graceful shutdown
            tmux send-keys -t "$SESSION_NAME" C-c
            sleep 2
            # Kill session if still running
            if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
                tmux kill-session -t "$SESSION_NAME"
            fi
            log_msg "Himiko is resting now... 💤"
        else
            log_msg "I'm not running... did you forget about me? 💔"
        fi
        ;;

    restart)
        echo "💉 Restarting Himiko... I'll be right back~ 💉"
        "$0" stop
        sleep 1
        "$0" start
        ;;

    status)
        if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
            echo "💕 Himiko is running in tmux session '$SESSION_NAME'~ 💕"
            echo "   To attach: tmux attach -t $SESSION_NAME"
        else
            echo "💔 Himiko is not running... I'm waiting for you to start me~ 💔"
        fi
        ;;

    attach)
        if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
            echo "💉 Connecting you to Himiko~ 💉"
            tmux attach -t "$SESSION_NAME"
        else
            echo "💔 I'm not running... start me first with: $0 start"
            exit 1
        fi
        ;;

    *)
        echo ""
        echo "  💉 Himiko Discord Bot - Session Manager 💉"
        echo ""
        echo "  Usage: $0 {start|stop|restart|status|attach}"
        echo ""
        echo "  Commands:"
        echo "    start   - Wake Himiko up in a new tmux session~"
        echo "    stop    - Let Himiko rest (stops the bot)"
        echo "    restart - Give Himiko a fresh start~"
        echo "    status  - Check if Himiko is running"
        echo "    attach  - Connect to Himiko's terminal session"
        echo ""
        exit 1
        ;;
esac
