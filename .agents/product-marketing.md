# Product Marketing Context

*Last updated: 2026-07-25*

## Product Overview
**One-liner:** Safe PocketBase access for AI coding agents.
**What it does:** pb-agent gives coding agents a bounded, auditable CLI for inspecting and changing PocketBase projects. Reads are structured for automation; mutations use immutable, expiring plans that humans can review before apply.
**Product category:** PocketBase developer tool, AI agent database gateway, PocketBase CLI.
**Product type:** Open-source command-line tool and agent skill.
**Business model:** Free and open source under Apache-2.0.

## Target Audience
**Target companies:** Individual developers, small product teams, agencies, and AI-native engineering teams building on PocketBase.
**Decision-makers:** Developers, technical founders, engineering leads, and platform engineers.
**Primary use case:** Let an AI coding agent inspect, test, and modify PocketBase during development without giving it an unrestricted shell around credentials or raw database access.
**Jobs to be done:**
- Diagnose PocketBase schema, records, rules, logs, and backups during development.
- Make reviewable data or schema changes through an agent.
- Keep credentials, production writes, and mutation evidence under explicit policy.
**Use cases:**
- Agent-assisted feature development and test setup.
- PocketBase schema and record debugging.
- Safe staging or production maintenance with scoped access grants.

## Personas
| Persona | Cares about | Challenge | Value we promise |
|---------|-------------|-----------|------------------|
| PocketBase developer | Shipping quickly | Agents need backend context and actions | One install gives agents a structured PocketBase workflow |
| Technical founder | Speed without incidents | Broad credentials make autonomous work risky | Plans, grants, redaction, and audit evidence |
| Engineering lead | Repeatability and review | Ad hoc scripts are hard to govern | One JSON-first contract across agents and environments |

## Problems & Pain Points
**Core problem:** AI coding agents need real PocketBase access to complete development work, but direct credentials plus arbitrary API or SQL access creates avoidable risk.
**Why alternatives fall short:**
- Generic MCP servers often expose direct operations without a durable plan/apply boundary.
- Raw HTTP and SDK scripts make every agent invent its own safety behavior.
- Human-only dashboards interrupt autonomous development and are difficult to audit.
**What it costs them:** Manual context switching, brittle one-off scripts, accidental writes, and low confidence in agent-made changes.
**Emotional tension:** Developers want agents to finish the job, but do not want to hand them an unbounded production credential.

## Competitive Landscape
**Direct:** PocketBase MCP servers that expose database operations as tools.
**Secondary:** PocketBase SDK scripts and generic REST clients.
**Indirect:** Manual PocketBase Dashboard operation.

## Differentiation
**Key differentiators:**
- Immutable, encrypted, expiring mutation plans.
- Environment-aware policy with production read-only by default.
- Optimistic concurrency checks before updates and deletes.
- OS-keychain credentials, recursive redaction, and metadata-only audit logs.
- JSON-first output plus a separate agent workflow skill.
**How we do it differently:** Safety is enforced by the capability kernel rather than left to prompts or agent discretion.
**Why that's better:** The same controls apply regardless of which coding agent invokes the CLI.
**Why customers choose us:** They need an agent to do useful PocketBase work and still want a reviewable boundary around changes.

## Objections
| Objection | Response |
|-----------|----------|
| Why not use an MCP server? | A local JSON-first CLI is portable across agent hosts and keeps the policy boundary independent of one protocol. MCP remains a possible thin adapter. |
| Does it store my password? | No. It uses a nonrenewable PocketBase superuser impersonation token stored in the OS credential manager. |
| Can an agent write to production? | Production is read-only by default. Mutations require a short-lived scoped access grant and an immutable plan. |

**Anti-persona:** Teams that need arbitrary SQL, raw HTTP passthrough, remote multi-tenant agent infrastructure, or unattended production writes.

## Switching Dynamics
**Push:** Repeated manual dashboard work and unsafe one-off scripts.
**Pull:** One agent-native command surface with built-in policy and evidence.
**Habit:** Existing SDK snippets and broad MCP tools appear quicker for the first operation.
**Anxiety:** Installing a new security-sensitive tool and trusting it with superuser access.

## Customer Language
**How they describe the problem:**
- "I want to give my coding agent access to PocketBase."
- "The agent should be able to do what it needs during development and testing."
- "How do I keep production safe?"
**How they describe us:**
- "A safe PocketBase CLI for AI agents."
- "Plan and review PocketBase changes before the agent applies them."
**Words to use:** safe, agent-first, PocketBase CLI, reviewable, plan then apply, local, open source.
**Words to avoid:** autonomous production, magic, seamless, enterprise-grade, fully capable.
**Glossary:**
| Term | Meaning |
|------|---------|
| Plan | Immutable encrypted description of a proposed mutation |
| Apply | Execute one previously created plan after policy checks |
| Access grant | Short-lived scope that permits a staging or production mutation |
| Connection | Named PocketBase URL, environment, and credential reference |

## Brand Voice
**Tone:** Calm, direct, technically precise.
**Style:** Short sentences, concrete claims, real commands, no hype.
**Personality:** Careful, capable, open, developer-friendly.

## Proof Points
**Metrics:** No adoption or outcome metrics claimed yet.
**Customers:** No customer logos claimed yet.
**Testimonials:** None claimed.
**Value themes:**
| Theme | Proof |
|-------|-------|
| Reviewable writes | Immutable 15-minute plans bound to one instance |
| Production restraint | Read-only default plus scoped access grants |
| Secret hygiene | OS keychain, stdin token input, recursive output redaction |
| Compatibility awareness | PocketBase 0.39.8 baseline and live capability probes |

## Goals
**Business goal:** Become the default open-source way developers connect AI coding agents to PocketBase.
**Conversion action:** Install pb-agent, then install the agent skill.
**Current metrics:** Pre-release; no public adoption baseline yet.
