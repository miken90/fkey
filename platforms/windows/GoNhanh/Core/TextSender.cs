using System.Runtime.InteropServices;

namespace GoNhanh.Core;

/// <summary>
/// Sends text to the active window using Unicode injection via SendInput API
/// Uses KEYEVENTF_UNICODE flag to inject characters directly as keyboard events
/// This preserves clipboard content and maintains uppercase state correctly
/// </summary>
public static class TextSender
{
    #region Win32 Constants

    private const uint INPUT_KEYBOARD = 1;
    private const uint KEYEVENTF_KEYUP = 0x0002;

    #endregion

    #region Win32 Imports

    [DllImport("user32.dll", SetLastError = true)]
    private static extern uint SendInput(uint nInputs, INPUT[] pInputs, int cbSize);

    #endregion

    #region Structures

    [StructLayout(LayoutKind.Sequential)]
    private struct INPUT
    {
        public uint type;
        public INPUTUNION u;
    }

    [StructLayout(LayoutKind.Explicit, Size = 32)]
    private struct INPUTUNION
    {
        [FieldOffset(0)] public KEYBDINPUT ki;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct KEYBDINPUT
    {
        public ushort wVk;
        public ushort wScan;
        public uint dwFlags;
        public uint time;
        public IntPtr dwExtraInfo;
    }

    #endregion

    /// <summary>
    /// Send text replacement using Unicode injection via SendInput API
    /// Preserves clipboard content and maintains uppercase state
    /// </summary>
    public static void SendText(string text, int backspaces)
    {
        if (string.IsNullOrEmpty(text) && backspaces == 0)
        {
            return;
        }

        var marker = KeyboardHook.GetInjectedKeyMarker();

        // Send all backspaces in one batch for better performance
        if (backspaces > 0)
        {
            SendBackspaces(backspaces, marker);
            Thread.Sleep(10); // Allow time for app to process backspaces before text injection
        }

        // Use Unicode injection for text insertion (preserves clipboard & uppercase)
        if (!string.IsNullOrEmpty(text))
        {
            SendUnicodeText(text, marker);
        }
    }

    /// <summary>
    /// Send multiple backspaces in a single batch
    /// </summary>
    private static void SendBackspaces(int count, IntPtr marker)
    {
        var inputs = new INPUT[count * 2];

        for (int i = 0; i < count; i++)
        {
            // Key down
            inputs[i * 2] = new INPUT
            {
                type = INPUT_KEYBOARD,
                u = new INPUTUNION
                {
                    ki = new KEYBDINPUT
                    {
                        wVk = KeyCodes.VK_BACK,
                        dwFlags = 0,
                        dwExtraInfo = marker
                    }
                }
            };

            // Key up
            inputs[i * 2 + 1] = new INPUT
            {
                type = INPUT_KEYBOARD,
                u = new INPUTUNION
                {
                    ki = new KEYBDINPUT
                    {
                        wVk = KeyCodes.VK_BACK,
                        dwFlags = KEYEVENTF_KEYUP,
                        dwExtraInfo = marker
                    }
                }
            };
        }

        SendInput((uint)inputs.Length, inputs, Marshal.SizeOf<INPUT>());
    }

    /// <summary>
    /// Send text using Unicode input via KEYEVENTF_UNICODE flag
    /// Injects characters directly as keyboard events (similar to macOS CGEvent.keyboardSetUnicodeString)
    /// Batches all characters in a single SendInput call for better reliability and performance
    /// </summary>
    private static void SendUnicodeText(string text, IntPtr marker)
    {
        const uint KEYEVENTF_UNICODE = 0x0004;

        // Batch all characters: each char needs 2 events (down + up)
        var inputs = new INPUT[text.Length * 2];
        int idx = 0;

        foreach (char c in text)
        {
            // Key down
            inputs[idx++] = new INPUT
            {
                type = INPUT_KEYBOARD,
                u = new INPUTUNION
                {
                    ki = new KEYBDINPUT
                    {
                        wVk = 0,
                        wScan = c,
                        dwFlags = KEYEVENTF_UNICODE,
                        dwExtraInfo = marker
                    }
                }
            };

            // Key up
            inputs[idx++] = new INPUT
            {
                type = INPUT_KEYBOARD,
                u = new INPUTUNION
                {
                    ki = new KEYBDINPUT
                    {
                        wVk = 0,
                        wScan = c,
                        dwFlags = KEYEVENTF_UNICODE | KEYEVENTF_KEYUP,
                        dwExtraInfo = marker
                    }
                }
            };
        }

        // Send all inputs in a single batch for atomic delivery
        SendInput((uint)inputs.Length, inputs, Marshal.SizeOf<INPUT>());
    }
}
