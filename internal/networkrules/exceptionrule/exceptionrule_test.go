package exceptionrule

import (
	"fmt"
	"testing"

	"github.com/anfragment/zen/internal/networkrules/rule"
)

func TestExceptionRule(t *testing.T) {
	t.Parallel()

	t.Run("'@@||page' should cancel '||page$document'", func(t *testing.T) {
		t.Parallel()

		filterName := "test"

		er := &ExceptionRule{
			RawRule:    "||example.com",
			FilterName: &filterName,
		}
		r := &rule.Rule{
			RawRule:    "||example.com$document",
			FilterName: &filterName,
		}
		r.ParseModifiers("document")

		want := true
		if got := er.Cancels(r); got != want {
			t.Errorf("'%s'.Cancels('%s') = %t, want %t", er.RawRule, r.RawRule, got, want)
		}
	})

	t.Run("'@@||page$document' should cancel '||page$document'", func(t *testing.T) {
		t.Parallel()

		filterName := "test"

		er := &ExceptionRule{
			RawRule:    "||example.com$document",
			FilterName: &filterName,
		}
		r := &rule.Rule{
			RawRule:    "||example.com$document",
			FilterName: &filterName,
		}
		r.ParseModifiers("document")
		er.ParseModifiers("document")

		want := true
		if got := er.Cancels(r); got != want {
			t.Errorf("'%s'.Cancels('%s') = %t, want %t", er.RawRule, r.RawRule, got, want)
		}
	})

	t.Run("'@@||page$document' should not cancel '||page'", func(t *testing.T) {
		t.Parallel()

		filterName := "test"

		er := &ExceptionRule{
			RawRule:    "||example.com^$document",
			FilterName: &filterName,
		}
		r := &rule.Rule{
			RawRule:    "||example.com",
			FilterName: &filterName,
		}
		er.ParseModifiers("document")

		want := false
		if got := er.Cancels(r); got != want {
			fmt.Println(got, want)
			t.Errorf("'%s'.Cancels('%s') = %t, want %t", er.RawRule, r.RawRule, got, want)
		}
	})
}
