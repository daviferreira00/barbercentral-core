package clientconfig

type ClientConfig struct {
	ClientID                 string  `json:"client_id" db:"client_id"`
	LogoURL                  *string `json:"logo_url,omitempty" db:"logo_url"`
	LogoCentral              *string `json:"logo_central,omitempty" db:"logo_central"`
	ColorPrimary             string  `json:"color_primary" db:"color_primary"`
	ColorSecondary           string  `json:"color_secondary" db:"color_secondary"`
	ColorButton              *string `json:"color_button,omitempty" db:"color_button"`
	BackgroundType           *string `json:"background_type,omitempty" db:"background_type"`
	FontFamily               string  `json:"font_family" db:"font_family"`
	Address                  *string `json:"address,omitempty" db:"address"`
	Neighborhood             *string `json:"neighborhood,omitempty" db:"neighborhood"`
	City                     *string `json:"city,omitempty" db:"city"`
	State                    *string `json:"state,omitempty" db:"state"`
	Phone                    *string `json:"phone,omitempty" db:"phone"`
	WhatsApp                 *string `json:"whatsapp,omitempty" db:"whatsapp"`
	Instagram                *string `json:"instagram,omitempty" db:"instagram"`
	Timezone                 string  `json:"timezone" db:"timezone"`
	CancellationPolicyHours int     `json:"cancellation_policy_hours" db:"cancellation_policy_hours"`
	BookingRequiresLogin    int     `json:"booking_requires_login" db:"booking_requires_login"` // 0=false, 1=true
	MinAdvanceHours         int     `json:"min_advance_hours" db:"min_advance_hours"`
	MaxAdvanceDays          int     `json:"max_advance_days" db:"max_advance_days"`
	IntervalBetweenMinutes  int     `json:"interval_between_minutes" db:"interval_between_minutes"`
	Active                   int     `json:"active" db:"active"`
}

type UpdateConfigRequest struct {
	LogoURL                  *string `json:"logo_url"`
	LogoCentral              *string `json:"logo_central"`
	ColorPrimary             string  `json:"color_primary"`
	ColorSecondary           string  `json:"color_secondary"`
	ColorButton              *string `json:"color_button"`
	BackgroundType           *string `json:"background_type"`
	FontFamily               string  `json:"font_family"`
	Address                  *string `json:"address"`
	Neighborhood             *string `json:"neighborhood"`
	City                     *string `json:"city"`
	State                    *string `json:"state"`
	Phone                    *string `json:"phone"`
	WhatsApp                 *string `json:"whatsapp"`
	Instagram                *string `json:"instagram"`
	Timezone                 string  `json:"timezone"`
	CancellationPolicyHours int     `json:"cancellation_policy_hours"`
	BookingRequiresLogin    int     `json:"booking_requires_login"`
	MinAdvanceHours         int     `json:"min_advance_hours"`
	MaxAdvanceDays          int     `json:"max_advance_days"`
	IntervalBetweenMinutes  int     `json:"interval_between_minutes"`
}

type PublicClientData struct {
	ClientConfig
	ClientName   string  `json:"client_name"`
	ClientSlug   string  `json:"client_slug"`
	CustomDomain *string `json:"custom_domain,omitempty"`
}
