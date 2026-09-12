# Production and cost flow

Confirmed order → production job → review Materials → start production → complete job.

The production workspace has Overview, Materials, and Outsourcing. Materials combines the plan, reservations, optional early usage/waste, and correction history. There is no separate consumption step.

Available stock is reserved from the order's saved material requirements. The reservation input is the single quantity editor; editing an order-managed reservation also updates its remaining material allowance. Additional material can be reserved below. Completion records the remaining allowance and outstanding allocations, subtracting usage already recorded. All inventory entries, job status, and any invoice cost adjustment commit together. A shortage prevents completion without partially recording materials. Repeating completion or reopening and completing an unchanged job does not duplicate usage. An order becomes Ready only after every item has a completed job.

For lower usage, reduce or release the reservation before completion. To record waste separately, use the optional early usage/waste form. Its slider and Max use the selected material's active reservation total for this job. A typed quantity can exceed reservations when additional free stock is available; stock reserved for other jobs remains protected. Quantities entered are additional usage, except when explicitly editing a previous record. Returns and corrections retain the original movements and post compensating entries with the original recorded value.

## Costs and margins

- The original order estimate remains the quoted snapshot. Quoting uses the configured pricing basis; inventory usage uses weighted-average inventory value.
- Recorded cost is net material usage plus waste plus outsourcing. Costs already incurred by cancelled jobs remain included until explicitly reversed.
- Expected cost includes recorded cost, the remaining material allowance, and the saved in-house machine, labor, service, fixed, and overhead estimates. Material and waste pricing components are replaced by the material allowance and recorded usage. These estimates scale with the in-house share when work is outsourced. Items without a production job retain their order estimate.
- Expected margin is the order total after discount minus expected cost. Margin percentage divides that amount by sales after discount, using decimal arithmetic. It is not markup and does not depend on customer payments.
- Exact inventory value is allocated proportionally when using stock; draining the pool leaves zero quantity and zero value. Reversals restore the exact cost originally deducted, even if the displayed unit rate was rounded.

Production jobs retain their cost breakdown in schema version 25. Existing jobs acquire the breakdown available from their saved order item during migration. Open jobs continue to follow order changes. For legacy records without component details, the unaccounted portion of the saved total estimate is retained instead of treating every estimated Rial as an additional cost.

## Accounting profit

Invoice posting recognizes revenue and the material/waste cost already recorded for the order. Later production usage, corrections, outsourcing stock returns, and job removal reconcile that invoice's material cost through additional balanced journal entries. They do not rewrite the original invoice posting. Voiding an invoice reverses its original entries and all subsequent production cost adjustments.

Outsourcing is already posted as an expense and is not posted again as inventory cost. Machine, labor, and overhead allowances are operational estimates: business profit uses the actual expenses recorded in accounting. Payments settle balances; they do not change job margin or create revenue a second time. Owner profit allocation continues to use posted fiscal-period profit.

The service sales report allocates invoice discounts proportionally, preserving every Rial, and includes recorded outsourcing along with material costs. Its contribution amount is before other operating expenses.

Existing completed jobs are not automatically charged for historical stock usage that was never recorded. Reopen and review their Materials plan before completing them under this flow. Existing journal entries are preserved; subsequent production actions create any required cost reconciliation on the action date, subject to fiscal-period posting rules.

## Verification

Regression tests cover automatic completion, repeated completion, shortages, partial outsourcing, expected margin after discount, invoices posted before production, cost returns, deletion adjustments, exact stock valuation, and service report discounts/outsourcing. The browser audit in `frontend/scripts/production-flow-audit.mjs` checks all three steps at 1024px and 1600px against a disposable demo database.
