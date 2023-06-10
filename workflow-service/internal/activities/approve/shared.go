package approve

var SignalChannels = struct {
	UPLOAD_CHANNEL  string
	PROCESS_CHANNEL string
	APPROVE_CHANNEL string
}{
	UPLOAD_CHANNEL:  "UPLOAD_CHANNEL",
	PROCESS_CHANNEL: "PROCESS_CHANNEL",
	APPROVE_CHANNEL: "APPROVE_CHANNEL",
}

var RouteTypes = struct {
	UPLOAD  string
	PROCESS string
	APPROVE string
}{
	UPLOAD:  "UPLOAD",
	PROCESS: "PROCESS",
	APPROVE: "APPROVE",
}

type RouteSignal struct {
	Route string
}

type UploadChannelSignal struct {
	Route string
}

type ProcessChannelSignal struct {
	Route   string
	Process ProcessTable
}

type ApproveChannelSignal struct {
	Route     string
	IsAprrove bool
}
