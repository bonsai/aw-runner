---
name: AW Self Organization
on:
  workflow_dispatch:
  schedule:
    - cron: '0 3 * * 1'
---

# AW Self Organization

Observe repositories, discover structure, propose new semantic artifacts, and
prepare validated generation work.

## Pipeline

1. OBSERVE repository metadata and evidence.
2. DISCOVER clusters, capabilities, and relationships.
3. PROPOSE domains, synapses, agents, and workflows.
4. GENERATE candidate YAML and repository manifests.
5. VALIDATE against ontology, policy, and human approval.
6. EVOLVE by feeding validated structure back into the next observation cycle.

Generation must remain proposal-first. Repository creation and ontology
declaration are separate approval steps.
