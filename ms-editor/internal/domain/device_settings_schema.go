package domain

import "encoding/json"

func ValidateDeviceSettings(deviceTypeName string, settings json.RawMessage) *ValidationError {
	settingsObject, objectValidationError := decodeJSONObject(settings)
	if objectValidationError != nil {
		return objectValidationError
	}

	validationFields := make([]FieldError, 0)

	switch deviceTypeName {
	case "modbus_rtu":
		requiredStringField(settingsObject, "serial_port", "settings.", &validationFields)

		baudRate, ok := requiredIntField(settingsObject, "baud_rate", "settings.", &validationFields)
		if ok && !isAllowedInt(baudRate, 1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.baud_rate",
				Message: "must be one of: 1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200",
			})
		}

		dataBits, ok := requiredIntField(settingsObject, "data_bits", "settings.", &validationFields)
		if ok && !isAllowedInt(dataBits, 7, 8) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.data_bits",
				Message: "must be one of: 7, 8",
			})
		}

		parity := requiredStringField(settingsObject, "parity", "settings.", &validationFields)
		if parity != "" && !isOneOf(parity, "none", "even", "odd") {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.parity",
				Message: `must be one of: "none", "even", "odd"`,
			})
		}

		stopBits, ok := requiredIntField(settingsObject, "stop_bits", "settings.", &validationFields)
		if ok && !isAllowedInt(stopBits, 1, 2) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.stop_bits",
				Message: "must be one of: 1, 2",
			})
		}

		slaveID, ok := requiredIntField(settingsObject, "slave_id", "settings.", &validationFields)
		if ok && (slaveID < 1 || slaveID > 247) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.slave_id",
				Message: "must be between 1 and 247",
			})
		}

		timeout, ok := requiredIntField(settingsObject, "timeout_ms", "settings.", &validationFields)
		if ok && (timeout < 100 || timeout > 30000) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.timeout_ms",
				Message: "must be between 100 and 30000",
			})
		}
	case "modbus_tcp":
		host := requiredStringField(settingsObject, "host", "settings.", &validationFields)
		if host != "" && !isValidHost(host) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.host",
				Message: "must be a valid IP address or hostname",
			})
		}

		port, ok := requiredIntField(settingsObject, "port", "settings.", &validationFields)
		if ok && (port < 1 || port > 65535) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.port",
				Message: "must be between 1 and 65535",
			})
		}

		slaveID, ok := requiredIntField(settingsObject, "slave_id", "settings.", &validationFields)
		if ok && (slaveID < 0 || slaveID > 247) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.slave_id",
				Message: "must be between 0 and 247",
			})
		}

		timeout, ok := requiredIntField(settingsObject, "timeout_ms", "settings.", &validationFields)
		if ok && (timeout < 100 || timeout > 30000) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.timeout_ms",
				Message: "must be between 100 and 30000",
			})
		}
	case "snmp_v1", "snmp_v2c":
		host := requiredStringField(settingsObject, "host", "settings.", &validationFields)
		if host != "" && !isValidHost(host) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.host",
				Message: "must be a valid IP address or hostname",
			})
		}

		port, ok := requiredIntField(settingsObject, "port", "settings.", &validationFields)
		if ok && (port < 1 || port > 65535) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.port",
				Message: "must be between 1 and 65535",
			})
		}

		requiredStringField(settingsObject, "community", "settings.", &validationFields)

		timeout, ok := requiredIntField(settingsObject, "timeout_ms", "settings.", &validationFields)
		if ok && (timeout < 100 || timeout > 30000) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.timeout_ms",
				Message: "must be between 100 and 30000",
			})
		}

		retryCount, ok := requiredIntField(settingsObject, "retry_count", "settings.", &validationFields)
		if ok && (retryCount < 0 || retryCount > 10) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.retry_count",
				Message: "must be between 0 and 10",
			})
		}
	case "snmp_v3":
		host := requiredStringField(settingsObject, "host", "settings.", &validationFields)
		if host != "" && !isValidHost(host) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.host",
				Message: "must be a valid IP address or hostname",
			})
		}

		port, ok := requiredIntField(settingsObject, "port", "settings.", &validationFields)
		if ok && (port < 1 || port > 65535) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.port",
				Message: "must be between 1 and 65535",
			})
		}

		requiredStringField(settingsObject, "security_name", "settings.", &validationFields)

		securityLevel := requiredStringField(settingsObject, "security_level", "settings.", &validationFields)
		if securityLevel != "" && !isOneOf(securityLevel, "noAuthNoPriv", "authNoPriv", "authPriv") {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.security_level",
				Message: `must be one of: "noAuthNoPriv", "authNoPriv", "authPriv"`,
			})
		}

		if securityLevel == "authNoPriv" || securityLevel == "authPriv" {
			authProtocol := requiredStringField(settingsObject, "auth_protocol", "settings.", &validationFields)
			if authProtocol != "" && !isOneOf(authProtocol, "MD5", "SHA") {
				validationFields = append(validationFields, FieldError{
					Field:   "settings.auth_protocol",
					Message: `must be one of: "MD5", "SHA"`,
				})
			}

			requiredStringField(settingsObject, "auth_password", "settings.", &validationFields)
		}

		if securityLevel == "authPriv" {
			privProtocol := requiredStringField(settingsObject, "priv_protocol", "settings.", &validationFields)
			if privProtocol != "" && !isOneOf(privProtocol, "DES", "AES") {
				validationFields = append(validationFields, FieldError{
					Field:   "settings.priv_protocol",
					Message: `must be one of: "DES", "AES"`,
				})
			}

			requiredStringField(settingsObject, "priv_password", "settings.", &validationFields)
		}

		timeout, ok := requiredIntField(settingsObject, "timeout_ms", "settings.", &validationFields)
		if ok && (timeout < 100 || timeout > 30000) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.timeout_ms",
				Message: "must be between 100 and 30000",
			})
		}

		retryCount, ok := requiredIntField(settingsObject, "retry_count", "settings.", &validationFields)
		if ok && (retryCount < 0 || retryCount > 10) {
			validationFields = append(validationFields, FieldError{
				Field:   "settings.retry_count",
				Message: "must be between 0 and 10",
			})
		}
	default:
		validationFields = append(validationFields, FieldError{
			Field:   "settings",
			Message: "unsupported device type",
		})
	}

	if len(validationFields) == 0 {
		return nil
	}

	return NewValidationError(validationFields...)
}

func isAllowedInt(value int, allowedValues ...int) bool {
	for _, allowedValue := range allowedValues {
		if value == allowedValue {
			return true
		}
	}

	return false
}
