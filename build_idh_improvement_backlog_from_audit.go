#!/bin/zsh

set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
REPORT_DIR="$REPO_DIR/reports"
AUDIT_REPORT="$REPORT_DIR/latest_income_policy_audit_report.md"
BACKLOG_REPORT="$REPORT_DIR/idh_improvement_backlog.md"
PLAN_90D="$REPORT_DIR/idh_90_day_execution_plan.md"

if [[ ! -f "$AUDIT_REPORT" ]]; then
  print -r -- "Audit report not found: $AUDIT_REPORT"
  print -r -- "Run audit_income_files_with_policy_tools.go first."
  exit 1
fi

TMP_ROWS="$(mktemp /tmp/idh_backlog_rows.XXXXXX)"
cleanup() { rm -f "$TMP_ROWS"; }
trap cleanup EXIT

extract_action_score() {
  local action="$1"
  case "$action" in
    APPROVE_TRACK) echo 4 ;;
    TRANSFORM_TO_PREVENTION) echo 3 ;;
    NEEDS_REWRITE) echo 2 ;;
    BLOCK_REVIEW) echo 0 ;;
    *) echo 1 ;;
  esac
}

extract_pillar() {
  local file_lc="$1"
  if print -r -- "$file_lc" | rg -q '(saude|health|sus|hospital|medical|sanitation|agua|water|disease|ozone|higiene)'; then
    echo "Saude"
  elif print -r -- "$file_lc" | rg -q '(educacao|education|mec|curriculum|school|escola|aprendiz|learning)'; then
    echo "Educacao"
  elif print -r -- "$file_lc" | rg -q '(renda|income|salario|salary|credit|credito|emprego|trabalho|cashflow|benefit|beneficio)'; then
    echo "Renda"
  else
    echo "Governanca"
  fi
}

extract_prevention_score() {
  local direction_lc="$1"
  local s=0
  if print -r -- "$direction_lc" | rg -q '(pilot|protoc|prevent|metric|indicador|audit|log)'; then
    s=$((s + 2))
  fi
  if print -r -- "$direction_lc" | rg -q '(block operational use|escalate|formal channel)'; then
    s=$((s - 1))
  fi
  echo "$s"
}

# Parse table rows: | `file` | `ACTION` | direction |
rg -n '^\| `' "$AUDIT_REPORT" \
  | sed -E 's/^[0-9]+://' \
  | while IFS= read -r line; do
      file="$(print -r -- "$line" | sed -E 's/^\| `([^`]+)` \| `([^`]+)` \| (.*) \|$/\1/')"
      action="$(print -r -- "$line" | sed -E 's/^\| `([^`]+)` \| `([^`]+)` \| (.*) \|$/\2/')"
      direction="$(print -r -- "$line" | sed -E 's/^\| `([^`]+)` \| `([^`]+)` \| (.*) \|$/\3/')"

      [[ -n "$file" && -n "$action" ]] || continue

      file_lc="$(print -r -- "$file" | tr '[:upper:]' '[:lower:]')"
      dir_lc="$(print -r -- "$direction" | tr '[:upper:]' '[:lower:]')"
      pillar="$(extract_pillar "$file_lc")"
      action_score="$(extract_action_score "$action")"
      prevention_score="$(extract_prevention_score "$dir_lc")"
      total_score=$((action_score + prevention_score))

      # Urgency: prioritize transform/approve; deprioritize block
      case "$action" in
        APPROVE_TRACK) urgency="Alta" ;;
        TRANSFORM_TO_PREVENTION) urgency="Alta" ;;
        NEEDS_REWRITE) urgency="Media" ;;
        BLOCK_REVIEW) urgency="Baixa" ;;
        *) urgency="Media" ;;
      esac

      # Next step
      case "$action" in
        APPROVE_TRACK) next_step="Escalar como piloto com KPI e trilha de auditoria." ;;
        TRANSFORM_TO_PREVENTION) next_step="Converter em protocolo preventivo com meta de 90 dias." ;;
        NEEDS_REWRITE) next_step="Reescrever com objetivo populacional, risco evitavel e metrica." ;;
        BLOCK_REVIEW) next_step="Revisao formal antes de uso operacional." ;;
        *) next_step="Revisar e classificar com evidencias." ;;
      esac

      print -r -- "$total_score|$urgency|$pillar|$action|$file|$next_step" >> "$TMP_ROWS"
    done

TOTAL_ITEMS="$(wc -l < "$TMP_ROWS" | tr -d ' ')"
TOP_ITEMS=25
if [[ "$TOTAL_ITEMS" -lt "$TOP_ITEMS" ]]; then
  TOP_ITEMS="$TOTAL_ITEMS"
fi

RUN_AT="$(date '+%Y-%m-%d %H:%M:%S %Z (%z)')"

{
  echo "# IDH Improvement Backlog (from income audit)"
  echo
  echo "- Generated at: $RUN_AT"
  echo "- Source report: \`reports/latest_income_policy_audit_report.md\`"
  echo "- Items parsed: $TOTAL_ITEMS"
  echo
  echo "## Prioritized backlog"
  echo
  echo "| Priority Score | Urgency | Pillar | Action | Source file | Next step |"
  echo "|---|---|---|---|---|---|"
  sort -t'|' -k1,1nr "$TMP_ROWS" | head -n "$TOP_ITEMS" | while IFS='|' read -r score urgency pillar action file next_step; do
    echo "| $score | $urgency | $pillar | \`$action\` | \`$file\` | $next_step |"
  done
  echo
  echo "## Pillar distribution (all parsed items)"
  echo
  total_saude="$(awk -F'|' '$3=="Saude"{c++} END{print c+0}' "$TMP_ROWS")"
  total_educacao="$(awk -F'|' '$3=="Educacao"{c++} END{print c+0}' "$TMP_ROWS")"
  total_renda="$(awk -F'|' '$3=="Renda"{c++} END{print c+0}' "$TMP_ROWS")"
  total_gov="$(awk -F'|' '$3=="Governanca"{c++} END{print c+0}' "$TMP_ROWS")"
  echo "- Saude: $total_saude"
  echo "- Educacao: $total_educacao"
  echo "- Renda: $total_renda"
  echo "- Governanca: $total_gov"
  echo
  echo "## Balanced shortlist by pillar (top 3 each, excluding BLOCK_REVIEW)"
  echo
  echo "| Pillar | Action | Source file | Next step |"
  echo "|---|---|---|---|"
  for p in Saude Educacao Renda Governanca; do
    sort -t'|' -k1,1nr "$TMP_ROWS" \
      | awk -F'|' -v pillar="$p" '$3==pillar && $4!="BLOCK_REVIEW"{print $0}' \
      | head -n 3 \
      | while IFS='|' read -r score urgency pillar action file next_step; do
          [[ -n "$file" ]] || continue
          echo "| $pillar | \`$action\` | \`$file\` | $next_step |"
        done
  done
  echo
  echo "## Rule"
  echo "Read full evidence before interpersonal escalation."
} > "$BACKLOG_REPORT"

# Build 90-day plan from top candidates by pillar and urgency
TOP_FILE_LIST="$(mktemp /tmp/idh_top_files.XXXXXX)"
BALANCED_TOP_FILE_LIST="$(mktemp /tmp/idh_balanced_top_files.XXXXXX)"
trap 'cleanup; rm -f "$TOP_FILE_LIST" "$BALANCED_TOP_FILE_LIST"' EXIT

sort -t'|' -k1,1nr "$TMP_ROWS" \
  | awk -F'|' '$2=="Alta" && $4!="BLOCK_REVIEW"{print $0}' \
  | head -n 15 \
  > "$TOP_FILE_LIST"

# Balanced execution list (max 3 per pillar, excluding blocked)
: > "$BALANCED_TOP_FILE_LIST"
for p in Saude Educacao Renda Governanca; do
  sort -t'|' -k1,1nr "$TMP_ROWS" \
    | awk -F'|' -v pillar="$p" '$3==pillar && $4!="BLOCK_REVIEW"{print $0}' \
    | head -n 3 \
    >> "$BALANCED_TOP_FILE_LIST"
done

{
  echo "# IDH 90-Day Execution Plan"
  echo
  echo "- Generated at: $RUN_AT"
  echo "- Strategy: prioritize high-urgency items with prevention and measurable outcomes."
  echo
  echo "## Phase 1 (Days 1-15) - Baseline and pilot setup"
  echo "- Validate top files and define one KPI per item."
  echo "- Assign owner + audit evidence source for each pilot."
  echo "- Open pilot cards using populational_issue_prevention_idea_form.md."
  echo
  echo "## Phase 2 (Days 16-45) - Pilot execution"
  echo "- Run small pilots in health, education, and income categories."
  echo "- Capture weekly metrics and adjust protocol."
  echo "- Keep fail-safe gate for legal and data protection constraints."
  echo
  echo "## Phase 3 (Days 46-90) - Scale or rewrite"
  echo "- Scale pilots with measurable gains."
  echo "- Rewrite low-performing proposals."
  echo "- Archive blocked items with rationale and evidence."
  echo
  echo "## Top pilot candidates (balanced by pillar)"
  echo
  echo "| Pillar | Action | Source file | KPI focus |"
  echo "|---|---|---|---|"
  while IFS='|' read -r score urgency pillar action file next_step; do
    case "$pillar" in
      Saude) kpi="Risk reduction and service continuity" ;;
      Educacao) kpi="Learning access and retention" ;;
      Renda) kpi="Income stability and employability" ;;
      *) kpi="Audit quality and response time" ;;
    esac
    echo "| $pillar | \`$action\` | \`$file\` | $kpi |"
  done < "$BALANCED_TOP_FILE_LIST"
  echo
  echo "## Governance"
  echo "- No payment or interpersonal escalation without full evidence review."
  echo "- Keep monthly re-audit with audit_income_files_with_policy_tools.go."
} > "$PLAN_90D"

print -r -- "IDH backlog generated:"
print -r -- "$BACKLOG_REPORT"
print -r -- "$PLAN_90D"
