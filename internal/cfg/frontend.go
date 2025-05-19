package cfg

// UpdatePolicyEnum defines the available update policies for the frontend.
// See /main.go for the usage.
var UpdatePolicyEnum = []struct {
	Value  UpdatePolicyType
	TSName string
}{
	{UpdatePolicyAutomatic, "AUTOMATIC"},
	{UpdatePolicyPrompt, "PROMPT"},
	{UpdatePolicyDisabled, "DISABLED"},
}
