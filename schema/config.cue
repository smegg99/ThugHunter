package schema

@go(config)

#Config: {
	db: {
		path: string | *"@{datadir:db/thughunter.db}" @go(Path)
	} @go(Db)
	imap: {
		host:               string | *"localhost"
		port:               int & >0 & <=65535 | *993
		catch_all_username: string | *"@{keyring?:IMAP_CATCH_ALL_USERNAME}" @go(CatchAllUsername)
		catch_all_password: string | *"@{keyring?:IMAP_CATCH_ALL_USERNAME}" @go(CatchAllPassword)
		mbox:               string | *"INBOX"
		use_tls:            bool | *true @go(UseTls)
	} @go(Imap)
  scraper: {
    custom_query_strings: [...string] | *[] @go(QueryStrings)
    endpoints: {
      home: string @go(HomeEndpoint)
      register: string @go(RegisterEndpoint)
      login: string @go(LoginEndpoint)
      search: string @go(SearchEndpoint)
      settings_billing_plan: string @go(SettingsBillingEndpoint)
    }
    agents: {
      templates: {
        email: string | *"spermokulka{{.ACCOUNT_ID}}@change.me" @go(EmailTemplate)
        password: string | *"{{.RANDOM_NONSENSE}}-{{.ACCOUNT_ID}}" @go(PasswordTemplate)
        organization: string | *"org{{.ACCOUNT_ID}}" @go(OrganizationTemplate)
        first_name: string | *"First{{.ACCOUNT_ID}}" @go(FirstNameTemplate)
        last_name: string | *"Last{{.ACCOUNT_ID}}" @go(LastNameTemplate)
      },
      max_agents: int & >0 | *10 @go(MaxAgents)
      max_register_retries: int & >=0 | *3 @go(MaxRegisterRetries)
    }
    browser_binary_path: string | *"/usr/bin/google-chrome-stable" @go(BrowserBinaryPath)
    virtual_display: bool | *true @go(VirtualDisplay)
    minimal_browser: bool | *false @go(MinimalBrowser)
  }
  scanner: {
    ping_mode: "strict" | "soft" | *"soft" @go(PingMode)
    icmp_ping: bool | *true @go(IcmpPing)
    reject_blank_screenshots: bool | *true @go(RejectBlankScreenshots)
    workers: {
      ping_timeout_seconds: int & >0 | *3 @go(PingTimeoutSeconds)
      connect_timeout_seconds: int & >0 | *10 @go(ConnectTimeoutSeconds)
      banner_timeout_seconds: int & >0 | *10 @go(BannerTimeoutSeconds)
      screenshot_timeout_seconds: int & >0 | *15 @go(ScreenshotTimeoutSeconds)
      screenshot_delay_seconds: int & >=0 | *1 @go(ScreenshotDelaySeconds)
      screenshot_pause_seconds: int & >=0 | *5 @go(ScreenshotPauseSeconds)
      screenshot_max_workers: int & >0 | *32 @go(ScreenshotMaxWorkers)
      max_workers: int & >0 | *2000 @go(MaxWorkers)
    }
    templates: {
      vnc_command:        string | *"remmina -c vnc://{{.IP}}:{{.PORT}}" @go(VncCommandTemplate)
      rdp_command:        string | *"xdg-open rdp://{{.IP}}:{{.PORT}}" @go(RdpCommandTemplate)
      spice_command:      string | *"remote-viewer spice://{{.IP}}:{{.PORT}}" @go(SpiceCommandTemplate)
      ssh_command:        string | *"xdg-open ssh://{{.IP}}:{{.PORT}}" @go(SshCommandTemplate)
      http_command:       string | *"xdg-open http://{{.IP}}:{{.PORT}}" @go(HttpCommandTemplate)
      https_command:      string | *"xdg-open https://{{.IP}}:{{.PORT}}" @go(HttpsCommandTemplate)
      screenshot_command: string | *"vncdo -s {{.IP}}::{{.PORT}} -i --timeout {{.TIMEOUT}} --delay {{.DELAY}} pause {{.PAUSE}} capture {{.OUTPUT}}" @go(ScreenshotCommandTemplate)
    }
  }
	logger:      #LoggerConfig
	preferences: #Preferences
}

#ThemeMode: "auto" | "light" | "dark" | "lightHighContrast" | "darkHighContrast"

#LocaleCode: "en" | "pl"

#AccentMode: "auto" | "custom"

#MemoryUnit: "auto" | "mb" | "mib"

#Preferences: {
	theme:          #ThemeMode | *"auto"
	language:       #LocaleCode | *"en"
	accent_mode:    #AccentMode | *"auto" @go(AccentMode)
	accent_color:   string | *"" @go(AccentColor)
	close_to_tray:  bool | *true @go(CloseToTray)
	memory_unit:    #MemoryUnit | *"auto" @go(MemoryUnit)
}

#LoggerConfig: {
	verbose:               bool | *false
	no_color:              bool | *false @go(NoColor)
	show_caller:           bool | *false @go(ShowCaller)
	dir:                   string | *""
	max_size_mb:           int & >0 | *10 @go(MaxSizeMb)
	max_backups:           int & >=0 | *5 @go(MaxBackups)
	max_age_days:          int & >0 | *30 @go(MaxAgeDays)
	log_name:              string | *"@{datadir:logs}" @go(LogName)
	frontend_console_log:  bool | *true @go(FrontendConsoleLog)
}
