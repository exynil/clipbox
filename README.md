<img src="logo.svg">

<br>

A powerful clipboard manager for Wayland with **rofi** integration, featuring multiple buffers, pinning, image preview, and password masking.

📖 **[Documentation & Wiki](https://github.com/exynil/clipbox/wiki)**

---

<https://github.com/user-attachments/assets/f07e9a18-24f6-4dbb-8eb4-cff41786c92b>

## Features

- 🗂️ **5 Independent Buffers** - Organize your clipboard history across workspaces
- 📌 **Pin Important Items** - Keep frequently used entries at your fingertips
- 🖼️ **Image Preview** - Automatic thumbnail generation for copied images (JPEG, PNG, GIF)
- 🔒 **Password Masking** - Smart detection and masking of sensitive passwords
- 🎨 **Customizable UI** - onfigure markers, colors, and other display details
- ⚡ **Fast operation** — Quick and responsive performance
- 🎯 **Rofi Integration** - Seamless keyboard-driven workflow

## Installation

### Dependencies

- `wayland`
- `wl-clipboard`

### Required Tools

- `rofi` - for interactive UI

#### From AUR (Arch Linux)

```bash
yay -S clipbox
```

#### Build from Source

```bash
git clone https://github.com/exynil/clipbox.git
cd clipbox
go build -o clipbox
sudo mv clipbox /usr/local/bin/
```

## Breaking Changes

Until we reach version 1 you should expect breaking changes from release to release. Watch the changelogs to learn about them.

We try to not introduce breaking changes that result in a definitive loss of data, but you should expect to have to redo your configuration from time to time.

---

**Star ⭐ this repo if you find it useful!**