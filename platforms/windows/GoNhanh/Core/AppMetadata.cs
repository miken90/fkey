using System.Reflection;

namespace GoNhanh.Core;

/// <summary>
/// Centralized app metadata - matches macOS AppMetadata.swift
/// All project metadata in one place for consistency
/// </summary>
public static class AppMetadata
{
    // App Info
    public static readonly string Name = "FKey";
    public static readonly string DisplayName = "FKey";
    public static readonly string Tagline = "Bộ gõ tiếng Việt hiệu suất cao";

    // Version
    public static string Version
    {
        get
        {
            var version = Assembly.GetExecutingAssembly().GetName().Version;
            return $"{version?.Major ?? 1}.{version?.Minor ?? 0}.{version?.Build ?? 0}";
        }
    }

    // Author
    public static readonly string Author = "Canh Nguyen";
    public static readonly string AuthorEmail = "nguyenminhcanh1990@gmail.com";
    public static readonly string AuthorLinkedin = "";

    // Links
    public static readonly string Website = "";
    public static readonly string Repository = "https://github.com/miken90/gonhanh.org";
    public static readonly string IssuesUrl = "https://github.com/miken90/gonhanh.org/issues";
    public static readonly string SpecialThanks = "Special thanks to Kha Phan <nhatkha1407@gmail.com>";

    // Legal
    public static readonly string Copyright = $"Copyright (c) 2025 {Author}. All rights reserved.";
    public static readonly string License = "GPL-3.0-or-later";

    // Tech
    public static readonly string TechStack = "Rust + WPF";
}

/// <summary>
/// Input method descriptions - matches macOS InputMode
/// </summary>
public static class InputMethodInfo
{
    public static string GetName(InputMethod method) => method switch
    {
        InputMethod.Telex => "Telex",
        InputMethod.VNI => "VNI",
        _ => "Unknown"
    };

    public static string GetShortName(InputMethod method) => method switch
    {
        InputMethod.Telex => "T",
        InputMethod.VNI => "V",
        _ => "?"
    };

    public static string GetDescription(InputMethod method) => method switch
    {
        InputMethod.Telex => "aw, ow, w, s, f, r, x, j",
        InputMethod.VNI => "a8, o9, 1-5",
        _ => ""
    };

    public static string GetFullDescription(InputMethod method) => method switch
    {
        InputMethod.Telex => $"Telex ({GetDescription(method)})",
        InputMethod.VNI => $"VNI ({GetDescription(method)})",
        _ => ""
    };
}
