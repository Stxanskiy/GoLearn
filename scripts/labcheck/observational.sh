#!/usr/bin/env bash
# Tasks whose check legitimately passes before the student acts.
#
# A lab step like «посмотри helm status» or «проверь, что Pod'ы запущены» asks the
# student to LOOK at state that the previous step already created. Its check has
# nothing of its own to verify, so run.sh would flag it WEAK — the same verdict it
# gives a genuinely broken check on an action task. Listing such a step here turns
# that into an explicit OBS verdict, which keeps WEAK meaningful.
#
# This list is NOT a way to silence a failing check. Add an entry only when the
# task text itself asks the student to observe or confirm, never when it asks them
# to change something. If in doubt, fix the check instead.
OBSERVATIONAL="
ch-helm-lab1#2
ch-helm-lab1#3
ch-helm-lab2#2
ch-helm-lab2#6
ch-helm-lab4#2
ch-helm-lab4#4
ch-helm-lab4#6
ch-helm-lab4#9
ch-ckad-lab14-debug-service-config#3
"
