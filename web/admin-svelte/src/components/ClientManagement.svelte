<script>
  import { onMount } from "svelte";

  let clients = $state([]);
  let dashboards = $state([]);
  let loading = $state(true);
  let error = $state("");

  async function apiGet(endpoint) {
    const r = await fetch("/api" + endpoint);
    if (!r.ok) throw new Error("HTTP " + r.status);
    return r.json();
  }
  async function apiPut(endpoint, body) {
    const r = await fetch("/api" + endpoint, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body)
    });
    if (!r.ok) throw new Error("HTTP " + r.status);
    return r.json();
  }

  async function load() {
    try {
      loading = true;
      const [c, d] = await Promise.all([apiGet("/clients"), apiGet("/dashboards")]);
      clients = c.clients || [];
      dashboards = d.dashboards || [];
      error = "";
    } catch (e) {
      error = "Failed to load data: " + e.message;
    } finally {
      loading = false;
    }
  }

  async function assignDashboard(clientId, dashboardId) {
    try {
      await apiPut(`/clients/${clientId}`, { dashboard_id: parseInt(dashboardId) });
      await load();
    } catch (e) {
      error = "Assign failed: " + e.message;
    }
  }

  onMount(load);
</script>

<div>
  <div class="card-header">
    <h2>Client Management</h2>
    <button class="btn secondary" onclick={load}>🔄 Refresh</button>
  </div>
  <div class="card-content">
    {#if loading}
      <div>Loading clients…</div>
    {:else}
      {#if error}<div class="status err">{error}</div>{/if}
      {#if clients.length === 0}
        <p>No clients have connected yet.</p>
      {:else}
        <table class="table">
          <thead>
            <tr>
              <th>IP Address</th>
              <th>Last Seen</th>
              <th>User Agent</th>
              <th>Assigned Dashboard</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
          {#each clients as c}
            <tr>
              <td>{c.ip_address}</td>
              <td>{new Date(c.last_seen).toLocaleString()}</td>
              <td style="max-width:260px; overflow:hidden; text-overflow:ellipsis;">{c.user_agent || "Unknown"}</td>
              <td>
                <select class="select" bind:value={c.assigned_dashboard_id} onchange={(e) => assignDashboard(c.id, e.currentTarget.value)} style="min-width: 180px;">
                  <option value="">Default Dashboard</option>
                  {#each dashboards as d}
                    <option value={d.id}>{d.name}</option>
                  {/each}
                </select>
              </td>
              <td><span class="status ok">Online</span></td>
            </tr>
          {/each}
          </tbody>
        </table>
      {/if}
    {/if}
  </div>
</div>
