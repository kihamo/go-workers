// Code generated by "stringer -type=EventId -trimprefix=EventId -output event_id_string.go"; DO NOT EDIT.

package workers

import "strconv"

const _EventId_name = "DispatcherStatusChangedWorkerAddWorkerRemoveWorkerExecuteStartWorkerExecuteStopWorkerStatusChangedTaskAddTaskRemoveTaskExecuteStartTaskExecuteStopTaskStatusChangedListenerAddListenerRemove"

var _EventId_index = [...]uint8{0, 23, 32, 44, 62, 79, 98, 105, 115, 131, 146, 163, 174, 188}

func (i EventId) String() string {
	if i < 0 || i >= EventId(len(_EventId_index)-1) {
		return "EventId(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EventId_name[_EventId_index[i]:_EventId_index[i+1]]
}