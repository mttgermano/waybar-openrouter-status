package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/mttgermano/waybar-openrouter-code/internal/ccusage"
)

func FormatNumber(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func FormatDuration(minutes int) string {
	if minutes <= 0 {
		return "0m"
	}

	hours, mins := minutes/60, minutes%60

	switch {
	case hours > 0 && mins > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}

func barString(pct int, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filled := (pct * width) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	color := "#619e6e" // Green hex
	if pct >= 80 {
		color = "#913326" // Red hex
	}

	return "<span foreground='" + color + "'>" + bar + "</span>"
}

func calculatePrefix(width int) string {
	return strings.Repeat(" ", width)
}

func FormatTooltip(data *ccusage.Data) string {
	var sb strings.Builder

	// HEADER ------------------------------------------------------
	headerString := fmt.Sprintf("OPENROUTER · <span foreground='#bd93f9'>"+"%s"+"</span>", data.LastModel)
	headerConst := len("<span foreground='#bd93f9'></span>")

	tabSize := headerConst
	dashSize := headerConst
	if len(headerString) > 2*headerConst {
		tabSize = 1
		dashSize = len(headerString) - headerConst + 2
	} else if len(headerString) == 2*headerConst {
		tabSize = 1
	} else {
		tabSize = (2*headerConst - len(headerString)) / 2
	}
	if (2*tabSize + len(headerString) - headerConst) <= dashSize {
		dashSize += 1
		tabSize += 1
	}

	prefixString := calculatePrefix(tabSize)
	fmt.Fprintf(&sb, "<b>%s</b>\n",
		prefixString+headerString)

	fmt.Fprintf(&sb, strings.Repeat("-", dashSize)+"\n")

	// requests ----------------------------------------------------
	fmt.Fprintf(&sb, "<b>\uf1d8 Requests:</b> %d\n",
		data.Entries)

	pct := data.Entries * 100 / 1000
	bar := barString(pct, 10)

	now := time.Now()
	tomorrow := time.Date(
		now.Year(),
		now.Month(),
		now.Day()+1,
		0, 0, 0, 0,
		now.Location(),
	)
	time_delta := int(tomorrow.Sub(now).Minutes())
	resetStr := strings.ToUpper(FormatDuration(time_delta))

	fmt.Fprintf(&sb, "%s RESETS · %s", bar, resetStr)

	// tokens ------------------------------------------------------
	// Today tokens
	fmt.Fprintf(&sb, "\n<b>TODAY   :</b> %s [%s  %s ]\n",
		FormatNumber(data.DailyTotalTokens),
		FormatNumber(data.DailyInputTokens),
		FormatNumber(data.DailyOutputTokens))

	// Session tokens with breakdown
	fmt.Fprintf(&sb, "<b>SESSION :</b> %s [%s  %s ]",
		FormatNumber(data.BlockTotalTokens),
		FormatNumber(data.BlockInputTokens),
		FormatNumber(data.BlockOutputTokens))

	return sb.String()
}
