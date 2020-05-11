# Traffic Policy Translation

See `design-docs/traffic_policy_translation_ir` for the relevant design doc.

To summarize, translation is done in a three-step process:

1. validate individual traffic policies in isolation
    - output: a `validationStatus` on each Traffic Policy
1. ensure that validated traffic policies from the previous step that have the same destination 
(or source) do not conflict with each other
    - output: populating the `validatedTrafficPolicies` field on Mesh Service statuses with the state of all the
relevant user-written Traffic Policies
1. finally, pull Traffic Policies that are both a) validated, and b) merge-able off of Mesh Service statuses and translate them
into the relevant mesh-specific configuration.
