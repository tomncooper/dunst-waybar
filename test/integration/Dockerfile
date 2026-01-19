# Base image with all dependencies for integration testing
FROM archlinux:latest

# Install all required packages in one layer
RUN pacman -Sy --noconfirm \
    dbus \
    dunst \
    libnotify \
    xorg-server-xvfb \
    && pacman -Scc --noconfirm

# Create necessary directories
RUN mkdir -p /run/dbus /tmp

# Set up for session bus usage
ENV DBUS_SESSION_BUS_ADDRESS=unix:path=/tmp/dbus-session
ENV DISPLAY=:99

# Mark as test image
LABEL test.image=dunst-waybar-integration

# Default command
CMD ["sleep", "infinity"]
