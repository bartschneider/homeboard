<script>
  import ClientManagement from "./components/ClientManagement.svelte";
  import WidgetBuilder from "./components/WidgetBuilder.svelte";
  import DashboardCanvas from "./components/DashboardCanvas.svelte";

  const tabs = [
    { id: "clients", label: "👥 Clients", component: ClientManagement },
    { id: "widgets", label: "🧩 Widget Builder", component: WidgetBuilder },
    { id: "dashboards", label: "📊 Dashboard Canvas", component: DashboardCanvas }
  ];

  let activeTab = $state("clients");
</script>

<div class="container">
  <div class="header">
    <h1>E‑Paper Dashboard Admin (Svelte 5)</h1>
    <p>Manage clients, create widgets, and design dashboards</p>
  </div>

  <div class="tabs">
    {#each tabs as t}
      <button class="tab {activeTab === t.id ? 'active' : ''}" onclick={() => activeTab = t.id}>
        {t.label}
      </button>
    {/each}
  </div>

  <div class="card">
    <div class="card-content">
      {#if activeTab === "clients"}
        <ClientManagement />
      {:else if activeTab === "widgets"}
        <WidgetBuilder />
      {:else}
        <DashboardCanvas />
      {/if}
    </div>
  </div>
</div>
