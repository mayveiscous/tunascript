Review the implementation of the IMUI library in `src/imui` and its integration within `src/interpreter/builtins.go`.

CRITICAL TOKEN CONSTRAINTS:

* Only inspect files located within `src/imui` and `src/interpreter/builtins.go`.
* Do not read files outside these paths.
* Do not attempt to understand the entire codebase.
* Focus exclusively on:

  * IMUI architecture
  * Widget lifecycle
  * Rendering flow
  * Window creation
  * Paint/update loop
  * Builtin registration and interpreter integration

After reviewing the implementation, propose a detailed rewrite plan that replaces WIN32 GDI rendering with a custom Software Renderer.

Assume the following architectural goals:

* No external rendering libraries.
* No Direct2D, Direct3D, OpenGL, Vulkan, SDL, Raylib, GLFW, Fyne, Gio, Wails, or other UI/rendering frameworks.
* Rendering should be performed entirely by CPU into a framebuffer owned by IMUI.
* WIN32 may still be used for:

  * Window creation
  * Message processing
  * Input handling
  * Presenting the final framebuffer to the window
* WIN32 GDI should not be responsible for drawing widgets, shapes, text, borders, backgrounds, or layout elements.

When analyzing the current implementation:

1. Identify every place where rendering occurs.
2. Identify every GDI dependency and its purpose.
3. Explain how each rendering operation would map to software rendering.
4. Explain which systems can remain unchanged.
5. Explain which systems should be refactored.

Design the proposed architecture around:

* A framebuffer:

  * Width
  * Height
  * Pixel buffer (`[]uint32` or equivalent)

* A renderer interface capable of:

  * Clear
  * SetPixel
  * FillRect
  * DrawRect
  * DrawLine
  * DrawText
  * DrawImage

* Future expansion for anti-aliasing.

* Widget code that emits drawing commands through the renderer rather than directly calling GDI.

Provide:

1. A high-level architecture diagram.
2. The rendering lifecycle before the rewrite.
3. The rendering lifecycle after the rewrite.
4. A migration plan broken into phases.
5. Concrete Go interfaces and structs that should be introduced.
6. Expected performance considerations and tradeoffs.
7. Potential challenges, especially:

   * Text rendering
   * Clipping
   * Alpha blending
   * Window presentation
   * Resizing

Do not generate implementation code unless necessary to illustrate an architectural recommendation. Focus on architecture, responsibilities, data flow, and migration strategy.