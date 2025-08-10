<script>
  import { onMount } from "svelte";

  let dashboards = $state([]);
  let widgets = $state([]);
  let selectedDashboard = $state(null);
  let dashboardWidgets = $state([]);
  let newDashboardName = $state("");
  let loading = $state(false);
  let error = $state("");
  let success = $state("");

  async function apiGet(endpoint) {
    const r = await fetch("/api" + endpoint);
    if (!r.ok) throw new Error("HTTP " + r.status);
    return r.json();
  }
  async function apiPost(endpoint, body) {
    const r = await fetch("/api" + endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body)
    });
    if (!r.ok) throw new Error("HTTP " + r.status);
    return r.json();
  }
  async function apiDelete(endpoint) {
    const r = await fetch("/api" + endpoint, { method: "DELETE" });
    if (!r.ok) throw new Error("HTTP " + r.status);
    return r.json();
  }

  async function loadData() {
    try {
      loading = true;
      const [d, w] = await Promise.all([apiGet("/dashboards"), apiGet("/widgets")]);
      dashboards = d.dashboards || [];
      widgets = w.widgets || [];
      error = "";
    } catch (e) {
      error = "Failed to load data: " + e.message;
    } finally {
      loading = false;
    }
  }

  async function loadDashboard(id) {
    try {
      const d = await apiGet(`/dashboards/${id}`);
      selectedDashboard = d;
      dashboardWidgets = d.widgets || [];
    } catch (e) {
      error = "Failed to load dashboard: " + e.message;
    }
  }

  async function createDashboard() {
    if (!newDashboardName.trim()) return;
    try {
      const d = await apiPost("/dashboards", {
        name: newDashboardName,
        description: "Dashboard created on " + new Date().toLocaleDateString(),
        is_default: false
      });
      newDashboardName = "";
      await loadData();
      selectedDashboard = d;
      dashboardWidgets = [];
      success = "Dashboard created.";
    } catch (e) {
      error = "Create failed: " + e.message;
    }
  }

  async function addWidgetToDashboard(widgetId) {
    if (!selectedDashboard) return;
    try {
      await apiPost(`/dashboards/${selectedDashboard.id}/widgets`, {
        widget_id: widgetId,
        display_order: (dashboardWidgets?.length || 0) + 1
      });
      await loadDashboard(selectedDashboard.id);
      success = "Widget added.";
    } catch (e) {
      error = "Add failed: " + e.message;
    }
  }

  async function removeWidgetFromDashboard(widgetId) {
    if (!selectedDashboard) return;
    try {
      await apiDelete(`/dashboards/${selectedDashboard.id}/widgets/${widgetId}`);
      await loadDashboard(selectedDashboard.id);
      success = "Widget removed.";
    } catch (e) {
      error = "Remove failed: " + e.message;
    }
  }

  onMount(loadData);
</script>

{#if loading}
  <div>Loading dashboard data…</div>
{:else}
  {#if error}<div class="status err">{error}</div>{/if}
  {#if success}<div class="status ok">{success}</div>{/if}

  <div class="grid" style="grid-template-columns: 1fr auto; align-items: end; margin-bottom: 12px;">
    <div>
      <label class="label">Select Dashboard</label>
      <select class="select" onchange={(e) => e.currentTarget.value ? loadDashboard(e.currentTarget.value) : selectedDashboard = null}>
        <option value="">Choose a dashboard…</option>
        {#each dashboards as d}
          <option value={d.id} selected={selectedDashboard?.id === d.id}>{d.name}{d.is_default ? " (Default)" : ""}</option>
        {/each}
      </select>
    </div>
    <div class="grid" style="grid-template-columns: 1fr auto; gap: 8px;">
      <input class="input" placeholder="New dashboard name" bind:value={newDashboardName}/>
      <button class="btn primary" onclick={createDashboard}>➕ Create</button>
    </div>
  </div>

  {#if selectedDashboard}
    <div class="grid" style="gap: 16px;">
      <div>
        <h3 style="margin-bottom: 8px;">Widget Library</h3>
        <div class="grid" style="grid-template-columns: repeat(auto-fill, minmax(220px,1fr));">
          {#each widgets.filter(w => w.enabled) as w}
            <button class="btn secondary" onclick={() => addWidgetToDashboard(w.id)} style="text-align:left;">
              <div style="font-weight:600;">{w.name}</div>
              <div style="color:#666; font-size:.9rem;">{w.template_type}</div>
              <div style="color:#999; font-size:.8rem; margin-top:4px;">{w.description}</div>
            </button>
          {/each}
        </div>
      </div>

      <div>
        <h3 style="margin-bottom: 8px;">Dashboard: {selectedDashboard.name}</h3>
        {#if (dashboardWidgets?.length || 0) === 0}
          <div style="padding: 24px; color:#86868b; border:1px dashed var(--border); border-radius: 8px; background:#f9f9fb;">
            No widgets yet. Click a widget from the library to add it here.
          </div>
        {:else}
          <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));">
            {#each dashboardWidgets as dw}
              <div class="card" style="padding: 12px;">
                <div class="card-header" style="padding: 8px 0; border-bottom: 1px solid var(--border);">
                  <strong>{dw.widget?.name}</strong>
                  <button class="btn danger" onclick={() => removeWidgetFromDashboard(dw.widget_id)} style="padding:6px 10px; font-size:.85rem;">✕</button>
                </div>
                <div style="color:#666; font-size:.9rem; margin-top:8px;">{dw.widget?.template_type}</div>
                <div style="color:#999; font-size:.8rem;">Order: {dw.display_order}</div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
{/if}
