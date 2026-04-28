package realtime

import (
	"context"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

func (connection *Connection) enqueueAlarm(event domain.AlarmEvent) {
	connection.enqueuePayload(BuildAlarmMessage(event), event.TagID.String())
}

func (connection *Connection) enqueueAlarmsSnapshot(records []domain.AlarmStateRecord) {
	connection.enqueuePayload(BuildAlarmsSnapshotMessage(records), TopicAlarms)
}

func (connection *Connection) enqueueAlarmsSnapshotAsync() {
	if connection.hub == nil {
		return
	}

	go func() {
		loadContext, cancelLoad := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelLoad()

		records, loadError := connection.hub.LoadActiveAlarmSnapshot(loadContext)
		if loadError != nil {
			connection.logger.Warn(
				"failed to load alarms snapshot",
				"method",
				"Connection.enqueueAlarmsSnapshotAsync",
				"session_id",
				connection.sessionID,
				"error",
				loadError,
			)
			return
		}

		connection.enqueueAlarmsSnapshot(records)
	}()
}
