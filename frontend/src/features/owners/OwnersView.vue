<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
import InlineAlert from '../../components/ui/InlineAlert.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import MasterDetail from '../../components/layout/MasterDetail.vue';
import FormGrid from '../../components/ui/FormGrid.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import DataTableRow from '../../components/ui/DataTableRow.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTable from '../../components/ui/DataTable.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import FormField from '../../components/ui/FormField.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref } from 'vue';
import { Plus, RotateCcw, Save, Trash2, UserRound, X } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import {
  ownersApi,
  type FiscalPeriodRecord,
  type OwnerRecord,
  type OwnerTransactionRecord,
} from '../../api/owners';
import { accountingApi, type FinancialAccountRecord } from '../../api/accounting';
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import { confirmAction, normalizeError } from '../../ui/feedback';
import SelectField from '../../components/ui/SelectField.vue';
type Tab =
  | 'Overview'
  | 'Owners'
  | 'Transactions'
  | 'Capital & Drawings'
  | 'Loans / Current Accounts'
  | 'Profit Allocation';
const props = defineProps<{ currencyUnit: 'Rial' | 'Toman' }>();
const emit = defineEmits<{ notify: [string] }>();
import {useOwnersWorkspace} from './useOwnersWorkspace'
const {busy,runAction,pageLoading,runLoad,tabs,tab,owners,transactions,periods,accounts,selectedOwner,selectedPeriod,error,editing,saving,ownerForm,shareForm,txForm,periodForm,currentOwner,selectedPeriodRow,filteredTransactions,ownerName,money,load,syncShares,loadTransactions,startOwner,saveOwner,saveShares,setActive,deleteOwner,postTransaction,createPeriod,previewPeriod,closePeriod,reverseTransaction}=useOwnersWorkspace(props,emit)
</script>
<template>
  <div class="min-w-0 space-y-3">
    <WorkspaceStickyStack
      ><WorkspaceHeader
        title="Owners & fiscal periods"
        eyebrow="Finance / ownership"
        description="Keep capital, current accounts, and period allocations authoritative in the ledger."
        ><button class="btn btn-primary" type="button" @click="startOwner">
          <Plus :size="16" /> New owner
        </button></WorkspaceHeader
      >
      <nav aria-label="Owner workspace">
        <button
          class="btn btn-sm"
          v-for="item in tabs"
          :key="item"
          type="button"
          :class="tab === item ? 'btn-primary' : 'btn-ghost'"
          @click="tab = item"
        >
          {{ item }}
        </button>
      </nav></WorkspaceStickyStack
    ><LoadingState v-if="pageLoading" label="Loading records…" /><div v-show="!pageLoading" class="space-y-4">
    <InlineAlert v-if="error" role="alert" class="flex flex-wrap items-center gap-2" tone="error"
      >{{ error }}
      <button class="btn btn-ghost" @click="error = ''" aria-label="Dismiss">
        <X :size="14" /></button
    ></InlineAlert>
    <section v-if="tab === 'Overview'" class="space-y-4">
<AppPanel title="Ownership overview" subtitle="Ownership and profit share are independent." flush><DataTable><thead><tr><th>Owner</th><th class="text-end">Ownership</th><th class="text-end">Profit share</th><th class="text-end">Current balance</th></tr></thead><tbody><DataTableRow v-for="o in owners" :key="o.id" interactive @activate="selectedOwner=o.id;tab='Owners';syncShares()"><DataTableCell><strong>{{o.name}}</strong></DataTableCell><DataTableCell numeric>{{o.ownershipBps/100}}%</DataTableCell><DataTableCell numeric>{{o.profitSharingBps/100}}%</DataTableCell><DataTableCell numeric>{{money(o.currentBalanceRial)}}</DataTableCell></DataTableRow></tbody></DataTable></AppPanel>
<AppPanel title="Fiscal periods" flush><DataTable><thead><tr><th>Period</th><th>From</th><th>To</th><th>Status</th><th class="text-end">Profit / loss</th></tr></thead><tbody><DataTableRow v-for="p in periods" :key="p.id" interactive @activate="selectedPeriod=p.id;tab='Profit Allocation'"><DataTableCell><strong>{{p.name}}</strong></DataTableCell><DataTableCell>{{formatDateTime(p.startDate)}}</DataTableCell><DataTableCell>{{formatDateTime(p.endDate)}}</DataTableCell><DataTableCell><StatusBadge :label="p.status" :tone="p.status==='Closed'?'slate':'amber'" /></DataTableCell><DataTableCell numeric>{{money(p.profitLossRial)}}</DataTableCell></DataTableRow></tbody></DataTable></AppPanel></section>
    <MasterDetail v-else-if="tab === 'Owners'"
      ><AppPanel
        title="Owner register"
        subtitle="Ownership and profit-sharing are independent fixed-scale percentages."
        ><div>
          <DataTable
            ><thead>
              <tr>
                <th>Owner</th>
                <th class="text-end">Ownership</th>
                <th class="text-end">Profit share</th>
                <th class="text-end">Current balance</th>
              </tr>
            </thead>
            <tbody>
              <DataTableRow
                v-for="o in owners"
                :key="o.id"
                :class="{ 'bg-base-300': selectedOwner === o.id }"
                @activate="
                  selectedOwner = o.id;
                  syncShares();
                "
                interactive
                ><DataTableCell
                  ><span class="block font-medium">{{ o.name }}</span
                  ><span class="block text-xs text-base-content/60">{{
                    o.active ? 'Active' : 'Archived'
                  }}</span></DataTableCell
                ><DataTableCell numeric>{{ o.ownershipBps / 100 }}%</DataTableCell
                ><DataTableCell numeric>{{ o.profitSharingBps / 100 }}%</DataTableCell
                ><DataTableCell numeric>{{
                  money(o.currentBalanceRial)
                }}</DataTableCell></DataTableRow
              >
            </tbody></DataTable
          >
        </div></AppPanel
      ><InspectorShell
        v-if="editing"
        title="New owner"
        subtitle="Percentages are stored as basis points, never floats."
        ><form @submit.prevent="saveOwner" class="min-w-0 space-y-3">
          <FormField class="gap-1"
            ><span>Name</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="ownerForm.name"
              required /></FormField
          ><FormGrid
            ><FormField class="gap-1"
              ><span>Ownership % (bps)</span
              ><AppInput
                class="input w-full min-w-0"
                v-model.number="ownerForm.ownershipBps"
                type="number"
                min="0"
                max="10000"
                required /></FormField
            ><FormField class="gap-1"
              ><span>Profit share % (bps)</span
              ><AppInput
                class="input w-full min-w-0"
                v-model.number="ownerForm.profitSharingBps"
                type="number"
                min="0"
                max="10000"
                required /></FormField></FormGrid
          ><FormField class="gap-1"
            ><span>Phone</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="ownerForm.phone" /></FormField
          ><FormField class="gap-1"
            ><span>Email</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="ownerForm.email"
              type="email" /></FormField
          ><FormField class="gap-1"
            ><span>Notes</span
            ><AppTextarea
              v-model="ownerForm.notes"
              rows="3"
            />
          </FormField>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-ghost" type="button" @click="editing = false">Cancel</button
            ><button class="btn btn-primary" :disabled="saving"><Save :size="15" /> Save</button>
          </div>
        </form></InspectorShell
      ><InspectorShell
        v-else-if="currentOwner"
        title="Owner inspector"
        subtitle="Posted history is immutable; corrections are reversals."
        ><div class="min-w-0 space-y-3">
          <div><UserRound :size="19" /></div>
          <div class="min-w-0 space-y-3">
            <h3 class="text-sm font-semibold">{{ currentOwner.name }}</h3>
            <p>
              {{ currentOwner.ownershipBps / 100 }}% ownership ·
              {{ currentOwner.profitSharingBps / 100 }}% profit share
            </p>
          </div>
        </div>
        <form @submit.prevent="saveShares" class="min-w-0 space-y-3">
          <FormGrid
            ><FormField class="gap-1"
              ><span>Ownership bps</span
              ><AppInput
                class="input w-full min-w-0"
                v-model.number="shareForm.ownershipBps"
                type="number"
                min="0"
                max="10000" /></FormField
            ><FormField class="gap-1"
              ><span>Profit share bps</span
              ><AppInput
                class="input w-full min-w-0"
                v-model.number="shareForm.profitSharingBps"
                type="number"
                min="0"
                max="10000" /></FormField></FormGrid
          ><button class="btn btn-primary"><Save :size="15" /> Save future shares</button>
        </form>
        <dl class="grid min-w-0 gap-2 text-sm">
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Capital</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(currentOwner.capitalContributedRial) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Drawings</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(currentOwner.drawingsRial) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Current / payable</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(currentOwner.currentBalanceRial) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Loans payable</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(currentOwner.loanPayableRial) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Loans receivable</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(currentOwner.loanReceivableRial) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Allocated P/L</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(currentOwner.allocatedProfitLossRial) }}
            </dd>
          </div>
        </dl>
        <div class="flex flex-wrap items-center gap-2">
          <button
            class="btn btn-ghost"
            @click="
              tab = 'Transactions';
              loadTransactions;
            "
          >
            View transactions</button
          ><button class="btn btn-ghost" @click="setActive" :disabled="busy">
            <RotateCcw :size="15" /> {{ currentOwner.active ? 'Archive' : 'Reactivate' }}</button
          ><button class="btn btn-ghost" @click="deleteOwner" :disabled="busy"><Trash2 :size="15" /> Delete</button>
        </div></InspectorShell
      ></MasterDetail
    >
    <section
      v-else-if="
        tab === 'Transactions' || tab === 'Capital & Drawings' || tab === 'Loans / Current Accounts'
      "
      class="min-w-0 space-y-4"
    >
      <AppPanel
        title="Post owner transaction"
        subtitle="Every action becomes one balanced integer-Rial journal entry."
        ><form @submit.prevent="postTransaction" class="min-w-0 space-y-3">
          <SelectField
            v-model="selectedOwner"
            label="Owner"
            :options="owners.map((owner) => ({ label: owner.name, value: owner.id }))"
          /><SelectField
            v-model="txForm.type"
            label="Type"
            :options="[
              { label: 'Capital contribution', value: 'capital_contribution' },
              { label: 'Drawing (not an expense)', value: 'drawing' },
              { label: 'Business expense paid personally', value: 'owner_paid_expense' },
              { label: 'Owner reimbursement', value: 'owner_reimbursement' },
              { label: 'Owner → business loan', value: 'loan_to_business' },
              { label: 'Business → owner loan', value: 'loan_from_business' },
              { label: 'Repay owner loan', value: 'loan_repayment_to_owner' },
              { label: 'Receive owner loan repayment', value: 'loan_repayment_from_owner' },
            ]"
          /><FormField class="gap-1"
            ><span>Amount ({{ currencyUnit }})</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="txForm.amountText"
              inputmode="decimal"
              required
              @blur="
                txForm.amountText = formatMoneyInput(
                  parseMoneyInput(txForm.amountText, currencyUnit) || 0,
                  currencyUnit,
                )
              " /></FormField
          ><SelectField
            v-if="txForm.type !== 'owner_paid_expense'"
            v-model="txForm.financialAccountId"
            label="Cash / bank account"
            :options="[
              { label: 'Select account', value: '' },
              ...accounts.map((account) => ({ label: account.name, value: account.id })),
            ]"
          /><FormField class="gap-1" v-else
            ><span>Expense category account</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="txForm.categoryAccountId"
              required /></FormField
          ><FormField class="gap-1"
            ><span>Description</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="txForm.description" /></FormField
          ><button class="btn btn-primary">Post transaction</button>
        </form></AppPanel
      ><AppPanel
        title="Owner transaction history"
        subtitle="Derived balances remain stable after reversals."
        ><div>
          <DataTable
            ><thead>
              <tr>
                <th>Number</th>
                <th>Owner</th>
                <th>Type</th>
                <th class="text-end">Amount</th>
                <th>Date</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="x in filteredTransactions" :key="x.id">
                <DataTableCell>{{ x.transactionNumber }}</DataTableCell
                ><DataTableCell>{{ ownerName(x.ownerId) }}</DataTableCell
                ><DataTableCell>{{ x.type }}</DataTableCell
                ><DataTableCell numeric>{{ money(x.amountRial) }}</DataTableCell
                ><DataTableCell>{{ formatDateTime(x.occurredAt) }}</DataTableCell
                ><DataTableCell
                  ><button
                    class="btn btn-ghost"
                    v-if="x.status === 'Posted'"
                    @click="reverseTransaction(x.id)"
                   :disabled="busy">
                    Reverse
                  </button></DataTableCell
                >
              </tr>
            </tbody></DataTable
          >
        </div></AppPanel
      >
    </section>
    <section v-else-if="tab === 'Profit Allocation'" class="min-w-0 space-y-4">
      <AppPanel
        title="Fiscal periods"
        subtitle="Close once; the profit-sharing snapshot is preserved forever."
      >
        <div class="flex flex-wrap items-center gap-2">
          <button
            v-for="p in periods"
            :key="p.id"
            :class="{ 'bg-base-300': selectedPeriod === p.id }"
            @click="selectedPeriod = p.id"
            class="btn btn-sm"
          >
            <span
              ><strong>{{ p.name }}</strong
              ><small class="block text-xs leading-5 text-base-content/60"
                >{{ formatDateTime(p.startDate) }} – {{ formatDateTime(p.endDate) }}</small
              ></span
            ><StatusBadge
              :label="p.status"
              :tone="p.status === 'Closed' ? 'slate' : 'amber'"
            /><b>{{ money(p.profitLossRial) }}</b>
          </button>
        </div>
        <form @submit.prevent="createPeriod" class="min-w-0 space-y-3">
          <h3 class="text-sm font-semibold">New period</h3>
          <FormField class="gap-1"
            ><span>Name</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="periodForm.name"
              required
              placeholder="1405" /></FormField
          ><FormGrid
            ><FormField class="gap-1"
              ><span>Start date</span><JalaliDatePicker v-model="periodForm.startDate" /></FormField
            ><FormField class="gap-1"
              ><span>End date</span
              ><JalaliDatePicker v-model="periodForm.endDate" /></FormField></FormGrid
          ><button class="btn btn-primary">Create period</button>
        </form>
      </AppPanel>
      <AppPanel
        v-if="selectedPeriodRow"
        title="Closing preview"
        subtitle="Revenue − COGS − expenses, from journal lines only."
        ><dl class="grid min-w-0 gap-2 text-sm">
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Revenue</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(selectedPeriodRow.revenueRial) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">COGS</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(selectedPeriodRow.cogsRial) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Expenses</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(selectedPeriodRow.expensesRial) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Profit / loss</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ money(selectedPeriodRow.profitLossRial) }}
            </dd>
          </div>
        </dl>
        <div class="flex flex-wrap items-center gap-2">
          <button class="btn btn-ghost" @click="previewPeriod" :disabled="busy">Refresh preview</button
          ><button
            class="btn btn-ghost"
            v-if="selectedPeriodRow.status === 'Open'"
            @click="closePeriod"
           :disabled="busy">
            Close period
          </button>
        </div>
        <h3 class="text-sm font-semibold">Owner allocation</h3>
        <div
          v-for="a in selectedPeriodRow.previewAllocations?.length
            ? selectedPeriodRow.previewAllocations
            : selectedPeriodRow.allocations"
          :key="a.ownerId"
          class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
        >
          <span>{{ ownerName(a.ownerId) }} · {{ a.profitSharingBps / 100 }}%</span
          ><b>{{ money(a.amountRial) }}</b>
        </div></AppPanel
      >
    </section>
  </div></div>
</template>
