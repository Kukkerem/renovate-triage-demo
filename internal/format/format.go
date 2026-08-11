// Package format renders values the way this provider surfaces them in
// status conditions and log lines.
//
// Casing is locale-aware rather than strings.Title, which is deprecated and
// wrong for non-ASCII input.
package format

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// titleCaser is stateless and safe for concurrent use, so it is built once
// instead of on every call.
var titleCaser = cases.Title(language.AmericanEnglish)

// PlanTitle renders a Cloud Foundry / BTP service plan name the way it is
// shown in provider status conditions, e.g. "standard" -> "Standard" and
// "trial-plan" -> "Trial-Plan".
func PlanTitle(name string) string {
	return titleCaser.String(name)
}
