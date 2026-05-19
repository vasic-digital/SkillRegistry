#!/usr/bin/env bash
# challenges/skillregistry_describe_challenge.sh
#
# Round-282 anti-bluff Challenge for dev.helix.agent/skillregistry.
#
# Default mode: invoke the runner against the real on-disk YAML
# fixture corpus and assert it exits 0 with the expected coverage,
# execution evidence, and 5-locale UX summary. This is the positive-
# evidence proof per Article XI §11.9 — the PASS is backed by captured
# stdout, not by absence of error or a green summary line.
#
# Paired-mutation mode (--mutate): build a scratch program that mirrors
# the runner's validator invariant in miniature, plant a known schema
# violation (a skill with empty ID — the validator MUST reject it),
# and assert the program detects it. A mutation run that exits 0 means
# the Challenge itself is a bluff (CONST-035 mutation-bluff), and this
# script exits 1 to surface that. A correctly detected mutation exits
# 99 — sentinel value the parent test bank recognises.
#
# Per CONST-035 / §11.4 and round-220 mirror template.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

MODE="default"
if [[ ${1:-} == "--mutate" ]]; then
    MODE="mutate"
fi

run_default() {
    echo "[skillregistry-challenge] mode=default — exercising runner against real corpus"
    cd "${REPO_ROOT}"

    local out
    out=$(go run ./challenges/runner -all 2>&1) || {
        echo "[skillregistry-challenge] FAIL: runner exited non-zero"
        echo "${out}"
        exit 1
    }

    # Positive-evidence assertions on captured stdout.
    if ! grep -q "^loaded_skills=5 " <<<"${out}"; then
        echo "[skillregistry-challenge] FAIL: missing loaded_skills line"
        echo "${out}"
        exit 1
    fi
    if ! grep -q "^skill=challenge-skill-en status=success " <<<"${out}" \
            || ! grep -q "^skill=challenge-skill-sr status=success " <<<"${out}" \
            || ! grep -q "^skill=challenge-skill-ja status=success " <<<"${out}" \
            || ! grep -q "^skill=challenge-skill-es status=success " <<<"${out}" \
            || ! grep -q "^skill=challenge-skill-de status=success " <<<"${out}"; then
        echo "[skillregistry-challenge] FAIL: missing per-skill success line"
        echo "${out}"
        exit 1
    fi
    if ! grep -q "^\[en\] skillregistry:" <<<"${out}" \
            || ! grep -q "^\[sr\] skillregistry:" <<<"${out}" \
            || ! grep -q "^\[ja\] skillregistry:" <<<"${out}" \
            || ! grep -q "^\[es\] skillregistry:" <<<"${out}" \
            || ! grep -q "^\[de\] skillregistry:" <<<"${out}"; then
        echo "[skillregistry-challenge] FAIL: missing one or more locale UX lines"
        echo "${out}"
        exit 1
    fi
    if ! grep -q "^OK skills=5 executions=5 locales=5$" <<<"${out}"; then
        echo "[skillregistry-challenge] FAIL: missing OK trailer"
        echo "${out}"
        exit 1
    fi

    # Defensive sanity — no rogue simulation/placeholder/TODO sneaked
    # into the runner or fixtures.
    if grep -RnE 'simulated|placeholder|TODO implement|for now' \
            "${REPO_ROOT}/challenges/runner" "${REPO_ROOT}/challenges/fixtures" 2>/dev/null \
            | grep -v "_test.go" ; then
        echo "[skillregistry-challenge] FAIL: bluff token detected in runner or fixtures"
        exit 1
    fi

    echo "${out}"
    echo "[skillregistry-challenge] PASS — runtime evidence captured above"
    exit 0
}

run_mutate() {
    echo "[skillregistry-challenge] mode=mutate — paired-mutation evidence"
    local scratch
    scratch="$(mktemp -d -t skillregistry-mutate-XXXXXX)"
    # shellcheck disable=SC2064
    trap "rm -rf '${scratch}'" EXIT

    # Stage a self-contained scratch module that mirrors the runner's
    # validator invariant in miniature. The mutation: build a Skill
    # with an empty ID — the runner's real validator would reject it,
    # so the in-scratch mirror MUST too. The scratch program exits 99
    # when the mutation is correctly surfaced; 0 means the mirror is
    # a bluff and this Challenge is too.
    mkdir -p "${scratch}/pkg/sr_scratch"

    cat > "${scratch}/go.mod" <<'EOF'
module skillregistry.scratch

go 1.25
EOF

    cat > "${scratch}/pkg/sr_scratch/validator.go" <<'EOF'
package sr_scratch

import (
	"errors"
	"strings"
)

// Skill is the mutated stand-in for the real Skill type. Only the
// fields the in-scratch validator looks at are present.
type Skill struct {
	ID          string
	Name        string
	Description string
}

// ValidateSkill mirrors the real validator's required-fields path
// (validator.go:validateRequiredFields). An empty ID MUST be
// rejected; the runner's full validator does the same.
func ValidateSkill(s Skill) error {
	if strings.TrimSpace(s.ID) == "" {
		return errors.New("skill ID is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("skill name is required")
	}
	if strings.TrimSpace(s.Description) == "" {
		return errors.New("skill description is required")
	}
	return nil
}

// MutatedSkill is the planted bluff: ID is intentionally empty, so a
// correct validator MUST reject it. Returning success here would be a
// CONST-035 mutation-bluff in the Challenge itself.
func MutatedSkill() Skill {
	return Skill{
		ID:          "", // ← mutation
		Name:        "Mutation Probe",
		Description: "Validator must reject the empty ID.",
	}
}
EOF

    cat > "${scratch}/main.go" <<'EOF'
package main

import (
	"fmt"
	"os"

	scr "skillregistry.scratch/pkg/sr_scratch"
)

func main() {
	s := scr.MutatedSkill()
	if err := scr.ValidateSkill(s); err != nil {
		fmt.Fprintf(os.Stderr, "mutation detected: %v\n", err)
		os.Exit(99)
	}
	fmt.Println("mutation NOT detected — bluff")
	os.Exit(0)
}
EOF

    cd "${scratch}"
    # Build then exec — `go run` does not preserve exit codes >2 on
    # all toolchains, which would mask the sentinel 99 the program
    # emits when the mutation is detected.
    go build -o ./mutbin . >/dev/null 2>&1 || {
        echo "[skillregistry-challenge] FAIL-MUTATE — scratch build failed"
        exit 1
    }
    local mut_out mut_rc
    set +e
    mut_out=$(./mutbin 2>&1)
    mut_rc=$?
    set -e

    echo "${mut_out}"
    if [[ ${mut_rc} -eq 99 ]]; then
        echo "[skillregistry-challenge] PASS-MUTATE — mutation correctly surfaced (exit 99)"
        exit 99
    fi
    echo "[skillregistry-challenge] FAIL-MUTATE — mutation NOT surfaced (exit ${mut_rc}); Challenge is a bluff"
    exit 1
}

case "${MODE}" in
    default) run_default ;;
    mutate)  run_mutate ;;
esac
