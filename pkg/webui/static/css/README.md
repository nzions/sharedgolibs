# GoogleEmu Web Components - Theme Structure

## Overview
The GoogleEmu Web Components Library has been restructured to use a modular theme architecture. Each theme is now in its own dedicated file for better maintainability and organization.

## File Structure
```
pkg/webui/static/css/
├── googleemu.css          # Main CSS file with base components and imports
├── themes/                # Theme directory containing all individual themes
│   ├── glassmorphism.css  # Glassmorphism theme (28 lines)
│   ├── professional.css   # Professional corporate theme (28 lines)
│   ├── puppies.css        # Fun and colorful theme (28 lines)
│   ├── weyland.css        # Weyland-Yutani retro-futuristic theme (170 lines)
│   ├── line-minimum.css   # Typewriter minimal theme (202 lines)
│   ├── waifu.css          # Kawaii anime aesthetic theme (378 lines)
│   └── hacker.css         # L33t h4ck3r Matrix theme (485 lines)
└── googleemu-old.css      # Backup of original monolithic file (1984 lines)
```

## Architecture Benefits

### ✅ Maintainability
- Each theme is isolated in its own file
- Easy to find and edit specific theme styles
- Clear separation of concerns

### ✅ Modularity  
- Themes can be developed independently
- Easy to add new themes without cluttering main file
- Individual themes can be included/excluded as needed

### ✅ Performance
- Selective loading possible (though currently all themes are imported)
- Easier to identify unused styles
- Better caching granularity

### ✅ Collaboration
- Multiple developers can work on different themes simultaneously
- Reduced merge conflicts
- Clear ownership of theme-specific code

## Theme Descriptions

### Glassmorphism (28 lines)
- Inspired by current portdash design
- Glass morphism effects with backdrop blur
- Purple gradient color scheme

### Professional (28 lines)  
- Clean corporate look
- Blue color scheme
- Minimal and business-appropriate

### Puppies (28 lines)
- Fun and colorful theme
- Warm orange/red gradient
- Playful and cheerful

### Weyland-Yutani (170 lines)
- Retro-futuristic corporate terminal theme
- Green monochrome color scheme
- Scanlines and terminal effects
- Extensive component customizations

### Line Minimum (202 lines)
- Typewriter minimal design
- Monospace typography
- Black and white color scheme
- Zero border radius for sharp edges

### Waifu/UwU (378 lines)
- Kawaii anime aesthetic
- Pink/purple pastel colors
- Hearts, sparkles, and cute animations
- Extensive kawaii-specific animations

### Hacker (485 lines)
- Authentic l33t h4ck3r Matrix style
- Green Matrix color scheme
- L33t speak text content
- Matrix rain, scanlines, and glitch effects
- Most complex theme with extensive animations

## Usage

The main `googleemu.css` file imports all themes automatically:
```css
@import url('./themes/glassmorphism.css');
@import url('./themes/professional.css');
@import url('./themes/hacker.css');
@import url('./themes/puppies.css');
@import url('./themes/weyland.css');
@import url('./themes/line-minimum.css');
@import url('./themes/waifu.css');
```

Themes are applied using the `data-theme` attribute:
```html
<body data-theme="hacker">
```

## Development Guidelines

### Adding New Themes
1. Create a new CSS file in `themes/` directory
2. Follow the naming convention: `theme-name.css`
3. Include proper header comment with theme description
4. Add the import statement to main `googleemu.css`
5. Update this documentation

### Modifying Existing Themes
1. Edit the specific theme file in `themes/` directory
2. Test changes across different components
3. Ensure theme-specific selectors use `[data-theme="theme-name"]`

### File Size Guidelines
- Simple themes (variables only): ~30 lines
- Medium themes (some components): ~200 lines  
- Complex themes (extensive styling): ~400+ lines

## Migration Notes
- Original monolithic file preserved as `googleemu-old.css`
- All functionality maintained
- No breaking changes to HTML or JavaScript
- All themes work exactly as before

## Version: 1.5.0
- Modular theme architecture
- 7 unique themes available
- 594 lines in main file (down from 1984)
- Individual theme files for better organization
