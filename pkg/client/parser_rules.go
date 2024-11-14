package client

import (
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"

	pb "github.com/cedi/icaltest/pkg/protos"
)

type RelabelConfig struct {
	Message   string `mapstructure:"message"`
	Important bool   `mapstructure:"important"`
}

type Rule struct {
	Name     string        `mapstructure:"name"`
	Key      string        `mapstructure:"key"`
	Contains []string      `mapstructure:"contains"`
	Skip     bool          `mapstructure:"skip"`
	Relabel  RelabelConfig `mapstructure:"relabelConfig"`
}

// Evaluate evaluates a rule against a pb.CalendarEntry and returns (bool, bool)
// where the first bool indicates if the rule was applied to this pb.CalendarEntry
// and the second bool indicates if this is a skip rule and the pb.CalendarEntry
// should be skipped
func (r *Rule) Evaluate(e *pb.CalendarEntry, zapLog *otelzap.Logger) (bool, bool) {

	var matchFieldValue string
	var matchFieldContains string
	match := false

	switch r.Key {
	case "*":
		fallthrough
	case "title":
		matchFieldValue = e.Title
		for _, contains := range r.Contains {
			if contains == "*" {
				match = true
			}

			if strings.Contains(e.Title, contains) {
				match = true
			}

			if match {
				matchFieldContains = contains
				break
			}
		}

	case "all_day":
		for _, contains := range r.Contains {
			matchFieldValue = strconv.FormatBool(e.AllDay)
			if contains == "*" {
				match = true
			}

			if strings.Contains(strconv.FormatBool(e.AllDay), contains) {
				match = true
			}

			if match {
				matchFieldContains = contains
				break
			}
		}
	}

	// The rule doesn't match
	if !match {
		return false, false
	}

	// perform the relabelings
	e.Message = r.Relabel.Message
	e.Important = r.Relabel.Important

	zapLog.Sugar().Debugw("Rule Evaluated", "rule_name", r.Name, "title", e.Title, "Field", matchFieldValue, "contains", matchFieldContains, "skip", r.Skip, "relabel_important", e.Important, "relabel_message", e.Message)

	return true, r.Skip
}

func parseRules() []Rule {
	rules := []Rule{}
	viper.UnmarshalKey("rules", &rules)
	return rules
}
