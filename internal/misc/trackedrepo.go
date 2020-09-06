package misc

// TrackedRepo identifies a git remote repository.
type TrackedRepo string

const RemoteOrigin = TrackedRepo("origin")
const RemoteUpstream = TrackedRepo("upstream")

var RecognizedRemotes = []TrackedRepo{RemoteUpstream, RemoteOrigin}
