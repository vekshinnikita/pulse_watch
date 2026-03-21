package errs

const (
	AlertRuleNotFoundErrorMessage       = "alert rule not found"
	AlertRuleDuplicateErrorMessage      = "rule with such parameters already exists."
	AlertRuleCheckThresholdErrorMessage = "threshold should be greater then 0"
	AlertRuleCheckIntervalErrorMessage  = "interval should be greater then 0"
)

type DuplicateAlertRuleError struct {
	Message string
}

func (e *DuplicateAlertRuleError) Error() string {
	return e.Message
}
