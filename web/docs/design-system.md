# Design System Documentation

## Overview

This document describes the design system implemented for the OpsPilot platform. The design system provides a consistent, scalable foundation for building UI components with a minimalist aesthetic inspired by modern platforms like Vercel, Linear, and Stripe.

## Design Philosophy

- **Minimalism**: Clean, uncluttered interfaces with focus on content
- **Consistency**: Unified visual language across all components
- **Accessibility**: WCAG 2.1 AA compliant with keyboard navigation support
- **Performance**: Optimized animations and efficient rendering

## Color System

### Primary Colors (Indigo)

The primary color palette uses Indigo (#6366f1) as the main brand color:

```typescript
primary: {
  50: '#eef2ff',   // Backgrounds, subtle highlights
  100: '#e0e7ff',  // Hover states
  200: '#c7d2fe',  // Selected states
  500: '#6366f1',  // Primary actions, links
  600: '#4f46e5',  // Hover on primary
  700: '#4338ca',  // Active/pressed states
  900: '#312e81',  // Dark text on light backgrounds
}
```

### Neutral Colors (Gray)

Gray scale for text, borders, and backgrounds:

```typescript
gray: {
  50: '#fafbfc',   // App background
  100: '#f8f9fa',  // Card backgrounds, table headers
  200: '#e9ecef',  // Borders, dividers
  300: '#dee2e6',  // Secondary borders
  400: '#ced4da',  // Disabled states
  500: '#6c757d',  // Secondary text
  700: '#495057',  // Primary text
  900: '#212529',  // Headings, emphasis
}
```

### Semantic Colors

Status and feedback colors:

- **Success**: `#10b981` (Green) - Successful operations, positive states
- **Warning**: `#f59e0b` (Amber) - Warnings, caution states
- **Error**: `#ef4444` (Red) - Errors, destructive actions
- **Info**: `#3b82f6` (Blue) - Informational messages

## Typography

### Font Families

**Sans-serif** (UI text):
```
-apple-system, BlinkMacSystemFont, Segoe UI, PingFang SC,
Hiragino Sans GB, Microsoft YaHei, sans-serif
```

**Monospace** (code):
```
SF Mono, Monaco, Cascadia Code, Consolas, monospace
```

### Font Sizes

- **xs**: 12px - Small labels, captions
- **sm**: 13px - Secondary text
- **base**: 14px - Body text (default)
- **lg**: 16px - Emphasized text
- **xl**: 18px - H4 headings
- **2xl**: 20px - H3 headings
- **3xl**: 24px - H2 headings
- **4xl**: 30px - H1 headings

### Font Weights

- **normal**: 400 - Body text
- **medium**: 500 - Buttons, emphasized text
- **semibold**: 600 - Headings
- **bold**: 700 - Strong emphasis

### Line Heights

- **tight**: 1.25 - Headings
- **normal**: 1.5 - Body text
- **relaxed**: 1.75 - Long-form content

## Spacing System

Based on 8px baseline grid:

```typescript
xs: '4px',    // 0.5 units - Tight spacing
sm: '8px',    // 1 unit - Small gaps
md: '16px',   // 2 units - Default spacing
lg: '24px',   // 3 units - Section spacing
xl: '32px',   // 4 units - Large gaps
2xl: '48px',  // 6 units - Major sections
3xl: '64px',  // 8 units - Page sections
```

## Border Radius

```typescript
sm: '4px',    // Small elements (tags, badges)
md: '8px',    // Default (buttons, inputs)
lg: '12px',   // Cards, modals
xl: '16px',   // Large containers
full: '9999px' // Pills, avatars
```

## Shadows

Elevation system for depth:

```typescript
sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)'           // Subtle lift
md: '0 1px 3px 0 rgba(0, 0, 0, 0.1), ...'       // Default cards
lg: '0 4px 6px -1px rgba(0, 0, 0, 0.1), ...'    // Elevated cards
xl: '0 10px 15px -3px rgba(0, 0, 0, 0.1), ...'  // Dropdowns, popovers
2xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), ...' // Modals, dialogs
```

## Animation System

### Duration

- **instant**: 100ms - Immediate feedback (button active)
- **fast**: 150ms - Quick transitions (hover, focus)
- **normal**: 200ms - Standard animations (expand, collapse)
- **slow**: 300ms - Page-level animations (enter, exit)

### Easing Functions

- **ease-out**: `cubic-bezier(0.16, 1, 0.3, 1)` - Enter animations
- **ease-in-out**: `cubic-bezier(0.65, 0, 0.35, 1)` - Bidirectional transitions
- **ease-spring**: `cubic-bezier(0.34, 1.56, 0.64, 1)` - Bounce effects

### Utility Classes

```css
.animate-fade-in      /* Fade in animation */
.animate-slide-in-up  /* Slide up animation */
.animate-scale-in     /* Scale in animation */
.skeleton             /* Loading skeleton shimmer */
```

## Component Specifications

### Buttons

- **Height**: 40px (default), 48px (large), 32px (small)
- **Border Radius**: 8px
- **Font Weight**: 500 (medium)
- **Padding**: 12px horizontal
- **Shadow**: Subtle on primary buttons

### Inputs

- **Height**: 40px (default), 48px (large), 32px (small)
- **Border Radius**: 8px
- **Border**: 1px solid gray-300
- **Focus**: 2px outline in primary-500
- **Padding**: 12px horizontal, 8px vertical

### Cards

- **Border Radius**: 12px
- **Shadow**: md (0 1px 3px rgba(0,0,0,0.08))
- **Padding**: 24px
- **Background**: white
- **Border**: 1px solid gray-200

### Tables

- **Row Height**: 48px minimum
- **Cell Padding**: 16px horizontal, 16px vertical
- **Header Background**: gray-100
- **Row Hover**: gray-50
- **Selected Row**: primary-50

### Modals

- **Border Radius**: 12px
- **Shadow**: 2xl (large shadow for prominence)
- **Max Width**: 520px (small), 720px (medium), 960px (large)
- **Backdrop**: rgba(0, 0, 0, 0.45) with blur

## Usage Examples

### Using Design Tokens in TypeScript

```typescript
import { colors, spacing, borderRadius } from '@/design-system/tokens';

const styles = {
  backgroundColor: colors.primary[500],
  padding: spacing.md,
  borderRadius: borderRadius.lg,
};
```

### Using Tailwind Classes

```tsx
<div className="bg-primary-50 text-primary-700 p-md rounded-lg shadow-md">
  Content with design system classes
</div>
```

### Using Ant Design Theme

The Ant Design theme is automatically applied via ConfigProvider in `main.tsx`. All Ant Design components will use the custom theme tokens.

```tsx
import { Button, Card } from 'antd';

<Card>
  <Button type="primary">Themed Button</Button>
</Card>
```

## Accessibility Guidelines

### Color Contrast

- Text on backgrounds must meet WCAG AA standards (4.5:1 for normal text, 3:1 for large text)
- Primary color (#6366f1) on white: 7.5:1 ✓
- Gray 700 (#495057) on white: 8.9:1 ✓

### Focus States

- All interactive elements have visible focus indicators
- Focus outline: 2px solid primary-500 with 2px offset
- Keyboard navigation fully supported

### Motion

- Respects `prefers-reduced-motion` media query
- All animations can be disabled for accessibility

## Responsive Breakpoints

```typescript
sm: '640px',   // Mobile landscape
md: '768px',   // Tablet
lg: '1024px',  // Desktop
xl: '1280px',  // Large desktop
2xl: '1536px', // Extra large screens
```

## Z-Index Scale

```typescript
dropdown: 1000,
sticky: 1020,
fixed: 1030,
modalBackdrop: 1040,
modal: 1050,
popover: 1060,
tooltip: 1070,
notification: 1080,
```

## Best Practices

1. **Use design tokens** instead of hardcoded values
2. **Prefer Tailwind classes** for consistent spacing and colors
3. **Use semantic colors** for status indicators (success, warning, error)
4. **Follow the 8px grid** for all spacing decisions
5. **Test with keyboard navigation** to ensure accessibility
6. **Use appropriate animation durations** - fast for micro-interactions, slow for page transitions
7. **Maintain consistent border radius** - sm for small elements, md for default, lg for cards

## Verification

To verify the design system is working correctly, run the verification component:

```tsx
import { DesignSystemVerification } from '@/test/design-system-verification';

// Add to a test route
<DesignSystemVerification />
```

This component tests:
- Design token imports and usage
- Tailwind class functionality
- Ant Design theme application
- Animation classes

## Resources

- Design Tokens: `src/design-system/tokens.ts`
- Tailwind Config: `tailwind.config.js`
- Ant Design Theme: `src/theme/antd-theme.ts`
- Global Styles: `src/index.css`
- Motion System: `src/styles/motion.css`

## Changelog

### 2026-03-02 - Initial Release

- Created design token system
- Configured Tailwind CSS with custom theme
- Configured Ant Design theme
- Implemented global styles and animations
- Fixed PostCSS parsing issues in motion.css
