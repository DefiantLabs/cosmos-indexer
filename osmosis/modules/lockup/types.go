package lockup

// Explicitly ignored messages for tx parsing purposes
const (
	MsgBeginUnlocking    = "/osmosis.lockup.MsgBeginUnlocking"
	MsgLockTokens        = "/osmosis.lockup.MsgLockTokens" // nolint:gosec
	MsgBeginUnlockingAll = "/osmosis.lockup.MsgBeginUnlockingAll"
	MsgUnlockPeriodLock  = "/osmosis.lockup.MsgUnlockPeriodLock"
	MsgUnlockTokens      = "/osmosis.lockup.MsgUnlockTokens" //nolint:gosec
)
