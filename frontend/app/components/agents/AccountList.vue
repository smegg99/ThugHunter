<!-- app/components/agents/AccountList.vue -->
<template>
  <v-table density="compact">
    <thead>
      <tr>
        <th class="text-left">{{ t('agents.accounts.email') }}</th>
        <th class="text-right">{{ t('agents.accounts.credits') }}</th>
        <th class="text-center">{{ t('agents.accounts.status') }}</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="account in accounts" :key="account.ID">
        <td class="text-body-2">{{ account.email }}</td>
        <td class="text-right text-body-2">{{ account.credits_amount }}</td>
        <td class="text-center">
          <v-chip :color="creditsExpired(account) ? 'error' : (account.credits_amount > 0 ? 'success' : 'warning')"
            size="x-small" variant="tonal" label>
            {{ creditsExpired(account) ? t('agents.accounts.expired') : (account.credits_amount > 0 ?
              t('agents.accounts.active') : t('agents.accounts.noCredits')) }}
          </v-chip>
        </td>
      </tr>
    </tbody>
  </v-table>
</template>

<script setup lang="ts">
import type { Account } from '~~bindings/smegg.me/thughunter/core/models/models.js'

const { t } = useI18n()

defineProps<{
  accounts: Account[]
}>()
</script>
