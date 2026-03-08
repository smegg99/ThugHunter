package schema

@go(config)

#Config: {
	db: {
		path: string | *"@{datadir:db/thughuntedb}"
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
      home: string @go(HomeUrl)
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
    }
    browser_binary_path: string | *"/usr/bin/google-chrome-stable" @go(BrowserBinaryPath)
    virtual_display: bool | *true @go(VirtualDisplay)
  }
	logger:      #LoggerConfig
	preferences: #Preferences
}

#ThemeMode: "auto" | "light" | "dark" | "lightHighContrast" | "darkHighContrast"

#LocaleCode: "en" | "pl"

#AccentMode: "auto" | "custom"

#Preferences: {
	theme:        #ThemeMode | *"auto"
	language:     #LocaleCode | *"en"
	accent_mode:  #AccentMode | *"auto" @go(AccentMode)
	accent_color: string | *"" @go(AccentColor)
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
