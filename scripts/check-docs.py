#!/usr/bin/env python3
"""Check the current planning documents and workflow without network access."""

from pathlib import Path
import re
import yaml


root = Path(__file__).resolve().parent.parent
documents = [root / "README.md", *sorted((root / "docs").glob("*.md"))]
required = {"prd.md", "implementation-plan.md", "development.md"}
errors = []

if not required.issubset({path.name for path in documents}):
    errors.append("Missing required planning/development documentation")

for path in documents:
    text = path.read_text(encoding="utf-8")
    if not text.endswith("\n"):
        errors.append(f"{path.relative_to(root)}: missing final newline")
    if text.count("```") % 2:
        errors.append(f"{path.relative_to(root)}: unclosed code fence")
    for number, line in enumerate(text.splitlines(), 1):
        if line.rstrip() != line:
            errors.append(f"{path.relative_to(root)}:{number}: trailing whitespace")
        if re.match(r"^(<{7}|={7}|>{7})( |$)", line):
            errors.append(f"{path.relative_to(root)}:{number}: conflict marker")
    for link in re.findall(r"\]\(([^\s)]+)(?:\s+\"[^\"]*\")?\)", text):
        if ":" in link or link.startswith("#"):
            continue
        target = path.parent / link.split("#", 1)[0]
        if not target.is_file():
            errors.append(f"{path.relative_to(root)}: missing local link {link}")

plan = (root / "docs/implementation-plan.md").read_text(encoding="utf-8")
for marker in [*(f"AC{i:02d}" for i in range(1, 10)),
               *(f"Q{i}:" for i in range(1, 4)),
               *(f"M{i}" for i in range(6))]:
    if marker not in plan:
        errors.append(f"Implementation plan is missing {marker}")

for path in sorted((root / ".github/workflows").glob("*.yml")):
    # BaseLoader retains GitHub's `on` key instead of YAML 1.1 boolean conversion.
    workflow = yaml.load(path.read_text(encoding="utf-8"), Loader=yaml.BaseLoader)
    if not isinstance(workflow, dict) or not {"on", "jobs"}.issubset(workflow):
        errors.append(f"{path.relative_to(root)}: missing workflow trigger/jobs")
        continue
    for job in workflow["jobs"].values():
        for step in job.get("steps", []):
            action = step.get("uses")
            if action and not re.fullmatch(r"[\w.-]+/[\w./-]+@[0-9a-f]{40}", action):
                errors.append(f"{path.relative_to(root)}: action is not commit-pinned: {action}")

if errors:
    raise SystemExit("\n".join(errors))
print("PASS: documentation links/hygiene, AC01–AC09, Q1–Q3, M0–M5, workflow YAML and action pins.")
