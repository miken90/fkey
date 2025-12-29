using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.Linq;
using GoNhanh.Core;

namespace GoNhanh.Services;

/// <summary>
/// Data structure for a shortcut with enabled state
/// </summary>
public record ShortcutData(string Replacement, bool IsEnabled);

/// <summary>
/// Manages user shortcuts (abbreviations like vn → Việt Nam)
/// Persists shortcuts to Registry with enabled/disabled state
/// </summary>
public class ShortcutsManager
{
    private const string RegistryPath = @"Software\GoNhanh\Shortcuts";
    private const string EnabledSuffix = "_enabled";
    private readonly Dictionary<string, ShortcutData> _shortcuts = new();

    /// <summary>
    /// Get all shortcuts as dictionary with data
    /// </summary>
    public IReadOnlyDictionary<string, ShortcutData> Shortcuts => _shortcuts;

    /// <summary>
    /// Load shortcuts from Registry and sync with Rust engine
    /// </summary>
    public void Load()
    {
        _shortcuts.Clear();
        RustBridge.ClearShortcuts();

        try
        {
            using var key = Registry.CurrentUser.OpenSubKey(RegistryPath);
            if (key == null) return;

            // Get all value names and filter out the _enabled suffix ones
            var valueNames = key.GetValueNames()
                .Where(n => !n.EndsWith(EnabledSuffix))
                .ToList();

            foreach (var valueName in valueNames)
            {
                var replacement = key.GetValue(valueName) as string;
                if (!string.IsNullOrEmpty(replacement))
                {
                    // Check if there's an enabled flag
                    var enabledValue = key.GetValue(valueName + EnabledSuffix);
                    var isEnabled = enabledValue == null || (int)enabledValue != 0;

                    _shortcuts[valueName] = new ShortcutData(replacement, isEnabled);

                    // Only add to Rust engine if enabled
                    if (isEnabled)
                    {
                        RustBridge.AddShortcut(valueName, replacement);
                    }
                }
            }

            System.Diagnostics.Debug.WriteLine($"[Shortcuts] Loaded {_shortcuts.Count} shortcuts");
        }
        catch (Exception ex)
        {
            System.Diagnostics.Debug.WriteLine($"[Shortcuts] Load error: {ex.Message}");
        }
    }

    /// <summary>
    /// Save all shortcuts to Registry
    /// </summary>
    public void Save()
    {
        try
        {
            // Delete old key and recreate to ensure clean state
            Registry.CurrentUser.DeleteSubKeyTree(RegistryPath, throwOnMissingSubKey: false);

            using var key = Registry.CurrentUser.CreateSubKey(RegistryPath);
            if (key == null) return;

            foreach (var (trigger, data) in _shortcuts)
            {
                key.SetValue(trigger, data.Replacement, RegistryValueKind.String);
                key.SetValue(trigger + EnabledSuffix, data.IsEnabled ? 1 : 0, RegistryValueKind.DWord);
            }

            System.Diagnostics.Debug.WriteLine($"[Shortcuts] Saved {_shortcuts.Count} shortcuts");
        }
        catch (Exception ex)
        {
            System.Diagnostics.Debug.WriteLine($"[Shortcuts] Save error: {ex.Message}");
        }
    }

    /// <summary>
    /// Sync all shortcuts to Rust engine (respecting enabled state)
    /// </summary>
    private void SyncToEngine()
    {
        RustBridge.ClearShortcuts();
        foreach (var (trigger, data) in _shortcuts)
        {
            if (data.IsEnabled)
            {
                RustBridge.AddShortcut(trigger, data.Replacement);
            }
        }
    }

    /// <summary>
    /// Add or update a shortcut (enabled by default)
    /// </summary>
    public void Add(string trigger, string replacement, bool isEnabled = true)
    {
        if (string.IsNullOrWhiteSpace(trigger) || string.IsNullOrWhiteSpace(replacement))
            return;

        _shortcuts[trigger] = new ShortcutData(replacement, isEnabled);

        // Update Rust engine
        if (isEnabled)
        {
            RustBridge.AddShortcut(trigger, replacement);
        }
        else
        {
            RustBridge.RemoveShortcut(trigger);
        }

        Save();
    }

    /// <summary>
    /// Update an existing shortcut's replacement and/or enabled state
    /// </summary>
    public void Update(string trigger, string replacement, bool isEnabled)
    {
        if (string.IsNullOrWhiteSpace(trigger))
            return;

        if (!_shortcuts.ContainsKey(trigger))
        {
            Add(trigger, replacement, isEnabled);
            return;
        }

        _shortcuts[trigger] = new ShortcutData(replacement, isEnabled);

        // Update Rust engine
        if (isEnabled)
        {
            RustBridge.AddShortcut(trigger, replacement);
        }
        else
        {
            RustBridge.RemoveShortcut(trigger);
        }

        Save();
    }

    /// <summary>
    /// Remove a shortcut
    /// </summary>
    public void Remove(string trigger)
    {
        if (_shortcuts.Remove(trigger))
        {
            RustBridge.RemoveShortcut(trigger);
            Save();
        }
    }

    /// <summary>
    /// Clear all shortcuts
    /// </summary>
    public void Clear()
    {
        _shortcuts.Clear();
        RustBridge.ClearShortcuts();
        Save();
    }

    /// <summary>
    /// Add default shortcuts (Vietnamese common abbreviations)
    /// </summary>
    public void LoadDefaults()
    {
        var defaults = new Dictionary<string, string>
        {
            { "vn", "Việt Nam" },
            { "hn", "Hà Nội" },
            { "hcm", "Hồ Chí Minh" },
            { "tphcm", "Thành phố Hồ Chí Minh" },
            { "ko", "không" },
            { "kg", "không" },
            { "dc", "được" },
            { "k", "không" },
            { "vs", "với" },
            { "ms", "mới" },
        };

        foreach (var (trigger, replacement) in defaults)
        {
            if (!_shortcuts.ContainsKey(trigger))
            {
                _shortcuts[trigger] = new ShortcutData(replacement, true);
                RustBridge.AddShortcut(trigger, replacement);
            }
        }

        Save();
    }
}
