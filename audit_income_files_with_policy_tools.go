#!/bin/zsh

set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
REPORT_DIR="$REPO_DIR/reports"
mkdir -p "$REPORT_DIR"

REPORT_FILE="$REPORT_DIR/latest_income_policy_audit_report.md"
IDEA_QUEUE_FILE="$REPORT_DIR/populational_issue_idea_queue.md"
FILE_LIST="$REPORT_DIR/income_policy_audit_file_list.txt"

NAME_MATCHES="$(mktemp /tmp/income_name_matches.XXXXXX)"
CONTENT_MATCHES="$(mktemp /tmp/income_content_matches.XXXXXX)"

cleanup() {
  rm -f "$NAME_MATCHES" "$CONTENT_MATCHES"
}
trap cleanup EXIT

# 1) Candidate files by name
rg --files "$REPO_DIR" \
  | rg -i '(income|renda|salary|salario|pagamento|payment|benefit|beneficio|cashflow|credito|credit|pro-labore|prolabore)' \
  | rg -v '/reports/|/\.git/' \
  > "$NAME_MATCHES" || true

# 2) Candidate files by content
rg -l -i -S \
  '(income|renda|salary|salario|pagamento|payment|benefit|beneficio|cashflow|credito|credit|pro-labore|prolabore|emprego|trabalho|aumento de renda|renda minima|renda estavel)' \
  "$REPO_DIR" \
  --glob '!**/.git/**' \
  --glob '!reports/**' \
  --glob '!**/*.png' \
  --glob '!**/*.jpg' \
  --glob '!**/*.jpeg' \
  --glob '!**/*.pdf' \
  > "$CONTENT_MATCHES" || true

cat "$NAME_MATCHES" "$CONTENT_MATCHES" | sort -u > "$FILE_LIST"

TOTAL_FILES=0
APPROVE_TRACK=0
TRANSFORM_TO_PREVENTION=0
BLOCK_REVIEW=0
NEEDS_REWRITE=0

RUN_AT="$(date '+%Y-%m-%d %H:%M:%S %Z (%z)')"

{
  echo "# Income File Policy Audit Report"
  echo
  echo "- Generated at: $RUN_AT"
  echo "- Repository: $REPO_DIR"
  echo "- Scope: files related to income/renda by name or content"
  echo
  echo "## Policy tools used"
  echo "- Prompt_Income_IDH_Guard_Bundle (approve/transform/block logic)"
  echo "- Politica_Leitura_Pagamento_Propostas_Bundle (objective criteria and payment gates)"
  echo "- incoming_income_audit_form.md (full-context and interpersonal gate)"
  echo "- populational_issue_prevention_idea_form.md (new ideas form)"
  echo
  echo "## File-by-file classification"
  echo
  echo "| File | Action | Prevention direction |"
  echo "|---|---|---|"
} > "$REPORT_FILE"

{
  echo "# Population Issue Idea Queue"
  echo
  echo "Use this queue with populational_issue_prevention_idea_form.md"
  echo "for files marked as TRANSFORM_TO_PREVENTION or NEEDS_REWRITE."
  echo
} > "$IDEA_QUEUE_FILE"

while IFS= read -r file; do
  [[ -n "$file" ]] || continue
  [[ -f "$file" ]] || continue

  TOTAL_FILES=$((TOTAL_FILES + 1))
  rel_path="${file#$REPO_DIR/}"

  content="$(sed -n '1,320p' "$file" 2>/dev/null | tr '[:upper:]' '[:lower:]')"

  action="NEEDS_REWRITE"
  prevention_direction="Add measurable prevention plan, signal, and metric."

  risk_hits="$({ print -r -- "$content" | rg -o '(fraud|golpe|scam|abuse|harass|violence|exploit|corrup|manipul)' 2>/dev/null || true; } | wc -l | tr -d ' ')"
  mitigation_hits="$({ print -r -- "$content" | rg -o '(preven|monitor|audit|auditoria|controle|integridade|compliance|governanca)' 2>/dev/null || true; } | wc -l | tr -d ' ')"
  evidence_hits="$({ print -r -- "$content" | rg -o '(metrica|indicador|evidenc|fonte|cronograma|custo|risco|entregaveis|score|justificativa)' 2>/dev/null || true; } | wc -l | tr -d ' ')"
  public_hits="$({ print -r -- "$content" | rg -o '(saude|health|educacao|education|renda|income|seguranca|safety|mobilidade|service|servico|trabalho|emprego|inclusao)' 2>/dev/null || true; } | wc -l | tr -d ' ')"

  if [[ "$risk_hits" -gt 0 && "$mitigation_hits" -eq 0 ]]; then
    action="BLOCK_REVIEW"
    prevention_direction="Block operational use, keep evidence, and escalate to formal channel."
    BLOCK_REVIEW=$((BLOCK_REVIEW + 1))
  elif [[ "$mitigation_hits" -gt 0 && "$evidence_hits" -gt 0 ]]; then
    action="APPROVE_TRACK"
    prevention_direction="Track indicators and keep continuous audit logs."
    APPROVE_TRACK=$((APPROVE_TRACK + 1))
  elif [[ "$public_hits" -gt 0 || "$mitigation_hits" -gt 0 ]]; then
    action="TRANSFORM_TO_PREVENTION"
    prevention_direction="Convert proposal to preventive protocol with pilot + impact metric."
    TRANSFORM_TO_PREVENTION=$((TRANSFORM_TO_PREVENTION + 1))

    {
      echo "## Idea candidate from \`$rel_path\`"
      echo "- Population issue:"
      echo "- Preventive action to test:"
      echo "- Pilot scope (90 days):"
      echo "- Primary metric:"
      echo "- Audit evidence required:"
      echo
    } >> "$IDEA_QUEUE_FILE"
  else
    NEEDS_REWRITE=$((NEEDS_REWRITE + 1))
    {
      echo "## Rewrite candidate from \`$rel_path\`"
      echo "- Rewrite into public-benefit prevention objective."
      echo "- Define affected population and expected risk reduction."
      echo "- Add legal/operational gate and one measurable metric."
      echo
    } >> "$IDEA_QUEUE_FILE"
  fi

  echo "| \`$rel_path\` | \`$action\` | $prevention_direction |" >> "$REPORT_FILE"
done < "$FILE_LIST"

{
  echo
  echo "## Summary"
  echo "- Total files audited: $TOTAL_FILES"
  echo "- APPROVE_TRACK: $APPROVE_TRACK"
  echo "- TRANSFORM_TO_PREVENTION: $TRANSFORM_TO_PREVENTION"
  echo "- BLOCK_REVIEW: $BLOCK_REVIEW"
  echo "- NEEDS_REWRITE: $NEEDS_REWRITE"
  echo
  echo "## Interpersonal gate"
  echo "Before any interpersonal approach, read full report + evidence and use:"
  echo "- incoming_income_audit_form.md"
  echo "- populational_issue_prevention_idea_form.md"
} >> "$REPORT_FILE"

print -r -- "Income policy audit completed."
print -r -- "Report: $REPORT_FILE"
print -r -- "Idea queue: $IDEA_QUEUE_FILE"
print -r -- "File list: $FILE_LIST"
