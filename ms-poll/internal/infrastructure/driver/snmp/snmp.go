package snmp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
	"github.com/gosnmp/gosnmp"
)

const (
	defaultSNMPPort      = 161
	defaultTimeoutMS     = 2000
	defaultRetries       = 1
	defaultCommunity     = "public"
	sysUpTimeOID         = "1.3.6.1.2.1.1.3.0"
	maxSNMPOIDsPerPacket = 25
)

type SNMPDriverFactory struct {
	version       gosnmp.SnmpVersion
	supportedType string
}

func NewVersionedFactory(version gosnmp.SnmpVersion) *SNMPDriverFactory {
	return &SNMPDriverFactory{
		version:       version,
		supportedType: supportedTypeForVersion(version),
	}
}

func (factory *SNMPDriverFactory) SupportedType() string {
	return factory.supportedType
}

func (factory *SNMPDriverFactory) Build(device *domain.DeviceSnapshot) (domain.Driver, error) {
	if device == nil {
		return nil, errors.New("device snapshot is nil")
	}

	client, buildError := buildSNMPClient(factory.version, device.Settings)
	if buildError != nil {
		return nil, fmt.Errorf("build snmp client: %w", buildError)
	}

	connectError := client.Connect()
	if connectError != nil {
		return nil, fmt.Errorf("connect snmp client: %w", connectError)
	}

	return &snmpDriver{
		client:       client,
		supportedType: factory.supportedType,
	}, nil
}

type snmpDriver struct {
	client        *gosnmp.GoSNMP
	supportedType string
}

func (driver *snmpDriver) Name() string {
	return driver.supportedType
}

func (driver *snmpDriver) Read(
	ctx context.Context,
	tags []*domain.TagSnapshot,
) []domain.Reading {
	if len(tags) == 0 {
		return nil
	}

	groupedTags := groupTagsByOID(tags)
	if len(groupedTags) == 0 {
		return badReadings(tags, errors.New("no valid snmp oids"))
	}

	orderedOIDs := make([]string, 0, len(groupedTags))
	for oid := range groupedTags {
		orderedOIDs = append(orderedOIDs, oid)
	}
	slices.Sort(orderedOIDs)

	readAt := time.Now().UTC()
	readings := make([]domain.Reading, len(tags))
	filled := make([]bool, len(tags))
	indexByTagID := make(map[string]int, len(tags))
	for index, tag := range tags {
		if tag == nil {
			continue
		}
		indexByTagID[tag.ID] = index
	}

	for start := 0; start < len(orderedOIDs); start += maxSNMPOIDsPerPacket {
		if ctxError := ctx.Err(); ctxError != nil {
			markChunkAsBad(
				readings,
				filled,
				indexByTagID,
				groupedTags,
				orderedOIDs[start:min(start+maxSNMPOIDsPerPacket, len(orderedOIDs))],
				ctxError,
				readAt,
			)
			continue
		}

		end := min(start+maxSNMPOIDsPerPacket, len(orderedOIDs))
		chunk := orderedOIDs[start:end]

		packet, getError := driver.client.Get(chunk)
		if getError != nil {
			markChunkAsBad(
				readings,
				filled,
				indexByTagID,
				groupedTags,
				chunk,
				getError,
				readAt,
			)
			continue
		}

		pduByOID := make(map[string]gosnmp.SnmpPDU, len(packet.Variables))
		for _, variable := range packet.Variables {
			pduByOID[normalizeOID(variable.Name)] = variable
		}

		for _, oid := range chunk {
			tagsByOID := groupedTags[oid]
			pdu, exists := pduByOID[oid]
			if !exists {
				markTagsAsBad(
					readings,
					filled,
					indexByTagID,
					tagsByOID,
					errors.New("snmp response does not contain requested oid"),
					readAt,
				)
				continue
			}

			for _, tag := range tagsByOID {
				if tag == nil {
					continue
				}
				tagIndex, exists := indexByTagID[tag.ID]
				if !exists {
					continue
				}

				value, decodeError := DecodePDU(pdu, tag.DataType)
				if decodeError != nil {
					readings[tagIndex] = newBadReading(tag, decodeError, readAt)
					filled[tagIndex] = true
					continue
				}

				readings[tagIndex] = domain.Reading{
					TagID:    tag.ID,
					DeviceID: tag.DeviceID,
					Value:    value,
					RawValue: value,
					Quality:  domain.QualityGood,
					ReadAt:   readAt,
				}
				filled[tagIndex] = true
			}
		}
	}

	for index, tag := range tags {
		if tag == nil || filled[index] {
			continue
		}

		readings[index] = newBadReading(
			tag,
			errors.New("reading is not available"),
			readAt,
		)
	}

	return readings
}

func (driver *snmpDriver) HealthCheck(ctx context.Context) error {
	if ctxError := ctx.Err(); ctxError != nil {
		return ctxError
	}

	_, getError := driver.client.Get([]string{sysUpTimeOID})
	if getError != nil {
		return fmt.Errorf("snmp health check failed: %w", getError)
	}

	return nil
}

func (driver *snmpDriver) Close() error {
	if driver.client == nil || driver.client.Conn == nil {
		return nil
	}

	closeError := driver.client.Conn.Close()
	if closeError != nil {
		return fmt.Errorf("close snmp connection: %w", closeError)
	}

	return nil
}

type v2Settings struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Community string `json:"community"`
	TimeoutMS int    `json:"timeout_ms"`
	Retries   int    `json:"retries"`
}

type v3Settings struct {
	Host             string `json:"host"`
	Port             int    `json:"port"`
	Username         string `json:"username"`
	AuthProtocol     string `json:"auth_protocol"`
	AuthPassphrase   string `json:"auth_passphrase"`
	PrivProtocol     string `json:"priv_protocol"`
	PrivPassphrase   string `json:"priv_passphrase"`
	SecurityLevel    string `json:"security_level"`
	TimeoutMS        int    `json:"timeout_ms"`
	Retries          int    `json:"retries"`
}

func buildSNMPClient(
	version gosnmp.SnmpVersion,
	rawSettings json.RawMessage,
) (*gosnmp.GoSNMP, error) {
	switch version {
	case gosnmp.Version1, gosnmp.Version2c:
		settings, parseError := parseV2Settings(rawSettings)
		if parseError != nil {
			return nil, parseError
		}

		return &gosnmp.GoSNMP{
			Target:    settings.Host,
			Port:      uint16(settings.Port),
			Community: settings.Community,
			Version:   version,
			Timeout:   time.Duration(settings.TimeoutMS) * time.Millisecond,
			Retries:   settings.Retries,
		}, nil
	case gosnmp.Version3:
		settings, parseError := parseV3Settings(rawSettings)
		if parseError != nil {
			return nil, parseError
		}

		msgFlags, securityParameters, securityError := buildV3Security(settings)
		if securityError != nil {
			return nil, securityError
		}

		return &gosnmp.GoSNMP{
			Target:             settings.Host,
			Port:               uint16(settings.Port),
			Version:            version,
			Timeout:            time.Duration(settings.TimeoutMS) * time.Millisecond,
			Retries:            settings.Retries,
			SecurityModel:      gosnmp.UserSecurityModel,
			MsgFlags:           msgFlags,
			SecurityParameters: securityParameters,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported snmp version: %d", version)
	}
}

func parseV2Settings(rawSettings json.RawMessage) (v2Settings, error) {
	settings := v2Settings{
		Port:      defaultSNMPPort,
		Community: defaultCommunity,
		TimeoutMS: defaultTimeoutMS,
		Retries:   defaultRetries,
	}

	if len(rawSettings) > 0 {
		if unmarshalError := json.Unmarshal(rawSettings, &settings); unmarshalError != nil {
			return v2Settings{}, fmt.Errorf("unmarshal snmp settings: %w", unmarshalError)
		}
	}

	settings.Host = strings.TrimSpace(settings.Host)
	if settings.Host == "" {
		return v2Settings{}, errors.New("host is required for snmp")
	}
	if settings.Port <= 0 {
		settings.Port = defaultSNMPPort
	}
	if settings.TimeoutMS <= 0 {
		settings.TimeoutMS = defaultTimeoutMS
	}
	if settings.Retries < 0 {
		settings.Retries = defaultRetries
	}
	settings.Community = strings.TrimSpace(settings.Community)
	if settings.Community == "" {
		settings.Community = defaultCommunity
	}

	return settings, nil
}

func parseV3Settings(rawSettings json.RawMessage) (v3Settings, error) {
	settings := v3Settings{
		Port:          defaultSNMPPort,
		TimeoutMS:     defaultTimeoutMS,
		Retries:       defaultRetries,
		SecurityLevel: "authPriv",
		AuthProtocol:  "SHA",
		PrivProtocol:  "AES",
	}

	if len(rawSettings) > 0 {
		if unmarshalError := json.Unmarshal(rawSettings, &settings); unmarshalError != nil {
			return v3Settings{}, fmt.Errorf("unmarshal snmp v3 settings: %w", unmarshalError)
		}
	}

	settings.Host = strings.TrimSpace(settings.Host)
	if settings.Host == "" {
		return v3Settings{}, errors.New("host is required for snmp_v3")
	}
	settings.Username = strings.TrimSpace(settings.Username)
	if settings.Username == "" {
		return v3Settings{}, errors.New("username is required for snmp_v3")
	}
	if settings.Port <= 0 {
		settings.Port = defaultSNMPPort
	}
	if settings.TimeoutMS <= 0 {
		settings.TimeoutMS = defaultTimeoutMS
	}
	if settings.Retries < 0 {
		settings.Retries = defaultRetries
	}
	settings.SecurityLevel = strings.TrimSpace(settings.SecurityLevel)
	if settings.SecurityLevel == "" {
		settings.SecurityLevel = "authPriv"
	}

	return settings, nil
}

func buildV3Security(
	settings v3Settings,
) (gosnmp.SnmpV3MsgFlags, *gosnmp.UsmSecurityParameters, error) {
	authProtocol, authProtocolError := parseAuthProtocol(settings.AuthProtocol)
	if authProtocolError != nil {
		return 0, nil, authProtocolError
	}
	privProtocol, privProtocolError := parsePrivProtocol(settings.PrivProtocol)
	if privProtocolError != nil {
		return 0, nil, privProtocolError
	}

	msgFlags, msgFlagsError := parseSecurityLevel(settings.SecurityLevel)
	if msgFlagsError != nil {
		return 0, nil, msgFlagsError
	}

	securityParameters := &gosnmp.UsmSecurityParameters{
		UserName:                 settings.Username,
		AuthenticationProtocol:   authProtocol,
		AuthenticationPassphrase: settings.AuthPassphrase,
		PrivacyProtocol:          privProtocol,
		PrivacyPassphrase:        settings.PrivPassphrase,
	}

	if msgFlags == gosnmp.AuthPriv || msgFlags == gosnmp.AuthNoPriv {
		if strings.TrimSpace(settings.AuthPassphrase) == "" {
			return 0, nil, errors.New("auth_passphrase is required for snmp_v3 auth mode")
		}
	}
	if msgFlags == gosnmp.AuthPriv && strings.TrimSpace(settings.PrivPassphrase) == "" {
		return 0, nil, errors.New("priv_passphrase is required for snmp_v3 authPriv mode")
	}

	return msgFlags, securityParameters, nil
}

func parseAuthProtocol(rawProtocol string) (gosnmp.SnmpV3AuthProtocol, error) {
	switch strings.ToUpper(strings.TrimSpace(rawProtocol)) {
	case "", "SHA":
		return gosnmp.SHA, nil
	case "MD5":
		return gosnmp.MD5, nil
	default:
		return gosnmp.NoAuth, fmt.Errorf("unsupported snmp auth_protocol: %s", rawProtocol)
	}
}

func parsePrivProtocol(rawProtocol string) (gosnmp.SnmpV3PrivProtocol, error) {
	switch strings.ToUpper(strings.TrimSpace(rawProtocol)) {
	case "", "AES":
		return gosnmp.AES, nil
	case "DES":
		return gosnmp.DES, nil
	default:
		return gosnmp.NoPriv, fmt.Errorf("unsupported snmp priv_protocol: %s", rawProtocol)
	}
}

func parseSecurityLevel(rawLevel string) (gosnmp.SnmpV3MsgFlags, error) {
	switch strings.ToLower(strings.TrimSpace(rawLevel)) {
	case "noauthnopriv":
		return gosnmp.NoAuthNoPriv, nil
	case "authnopriv":
		return gosnmp.AuthNoPriv, nil
	case "authpriv":
		return gosnmp.AuthPriv, nil
	default:
		return 0, fmt.Errorf("unsupported snmp security_level: %s", rawLevel)
	}
}

func supportedTypeForVersion(version gosnmp.SnmpVersion) string {
	switch version {
	case gosnmp.Version1:
		return "snmp_v1"
	case gosnmp.Version2c:
		return "snmp_v2c"
	case gosnmp.Version3:
		return "snmp_v3"
	default:
		return ""
	}
}

func groupTagsByOID(tags []*domain.TagSnapshot) map[string][]*domain.TagSnapshot {
	grouped := make(map[string][]*domain.TagSnapshot)

	for _, tag := range tags {
		if tag == nil {
			continue
		}

		oid, parseError := parseTagOID(tag.Address)
		if parseError != nil {
			continue
		}

		grouped[oid] = append(grouped[oid], tag)
	}

	return grouped
}

func parseTagOID(rawAddress json.RawMessage) (string, error) {
	var payload struct {
		OID string `json:"oid"`
	}

	if unmarshalError := json.Unmarshal(rawAddress, &payload); unmarshalError != nil {
		return "", fmt.Errorf("parse snmp tag address: %w", unmarshalError)
	}

	normalizedOID := normalizeOID(payload.OID)
	if normalizedOID == "" {
		return "", errors.New("snmp oid is required")
	}

	return normalizedOID, nil
}

func normalizeOID(oid string) string {
	return strings.TrimPrefix(strings.TrimSpace(oid), ".")
}

func markChunkAsBad(
	readings []domain.Reading,
	filled []bool,
	indexByTagID map[string]int,
	groupedTags map[string][]*domain.TagSnapshot,
	chunk []string,
	readError error,
	readAt time.Time,
) {
	for _, oid := range chunk {
		markTagsAsBad(
			readings,
			filled,
			indexByTagID,
			groupedTags[oid],
			readError,
			readAt,
		)
	}
}

func markTagsAsBad(
	readings []domain.Reading,
	filled []bool,
	indexByTagID map[string]int,
	tags []*domain.TagSnapshot,
	readError error,
	readAt time.Time,
) {
	for _, tag := range tags {
		if tag == nil {
			continue
		}

		tagIndex, exists := indexByTagID[tag.ID]
		if !exists {
			continue
		}

		readings[tagIndex] = newBadReading(tag, readError, readAt)
		filled[tagIndex] = true
	}
}

func badReadings(tags []*domain.TagSnapshot, readError error) []domain.Reading {
	readings := make([]domain.Reading, 0, len(tags))
	readAt := time.Now().UTC()
	for _, tag := range tags {
		readings = append(readings, newBadReading(tag, readError, readAt))
	}
	return readings
}

func newBadReading(
	tag *domain.TagSnapshot,
	readError error,
	readAt time.Time,
) domain.Reading {
	if tag == nil {
		return domain.Reading{
			Quality: domain.QualityBad,
			Error:   "tag is nil",
			ReadAt:  readAt,
		}
	}

	errorMessage := "snmp read failed"
	if readError != nil {
		errorMessage = readError.Error()
	}

	return domain.Reading{
		TagID:    tag.ID,
		DeviceID: tag.DeviceID,
		Quality:  domain.QualityBad,
		Error:    errorMessage,
		ReadAt:   readAt,
	}
}
