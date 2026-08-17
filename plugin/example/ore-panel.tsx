import { plugin, mine, type Block, type Bot } from "graft"

type Group = { label: string; ores: Block[] }

const groups: Group[] = [
  {
    label: "comuns",
    ores: ["coal_ore", "copper_ore", "iron_ore", "redstone_ore", "lapis_ore"],
  },
  {
    label: "profundos",
    ores: [
      "deepslate_coal_ore", "deepslate_copper_ore", "deepslate_iron_ore",
      "deepslate_gold_ore", "deepslate_redstone_ore", "deepslate_lapis_ore",
      "deepslate_diamond_ore", "deepslate_emerald_ore",
    ],
  },
  {
    label: "raros",
    ores: ["gold_ore", "diamond_ore", "emerald_ore", "ancient_debris"],
  },
  {
    label: "nether",
    ores: ["nether_quartz_ore", "nether_gold_ore"],
  },
]

const named = (ore: string) => ore.replace(/_ore$/, "").replace(/_/g, " ")

let tab = "minerar"
let raio = 32
let manterInventario = true
let voltarAoBau = false

const style = `
:root {
  --glass: rgba(22,26,38,.82);
  --line: rgba(255,255,255,.09);
  --dim: rgba(233,237,245,.45);
  --accent: #7dd3fc;
}

.scrim {
  position: absolute;
  top: 0; right: 0; bottom: 0; left: 0;
  background: rgba(6,8,14,.35);
}

.shell {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%,-50%);
  width: 460px;
  height: 520px;
  display: flex;
  flex-direction: column;
  border-radius: 18px;
  overflow: hidden;
  background:
    linear-gradient(160deg, rgba(255,255,255,.07), rgba(255,255,255,0) 42%),
    var(--glass);
  box-shadow:
    0 40px 90px rgba(0,0,0,.55),
    0 2px 0 rgba(255,255,255,.10) inset,
    0 0 0 1px rgba(255,255,255,.08);
}

header {
  display: flex;
  align-items: center;
  padding: 18px 20px 14px;
  border-bottom: 1px solid var(--line);
}
header h1 {
  margin: 0 0 0 12px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: .16em;
  text-transform: uppercase;
  color: var(--dim);
  flex: 1;
}
header .who {
  font-size: 12px;
  color: var(--dim);
}

nav {
  display: flex;
  padding: 12px 16px 0;
}
nav button + button { margin-left: 6px }
nav button {
  flex: 1;
  padding: 9px 0;
  border: 0;
  border-radius: 9px;
  background: transparent;
  color: var(--dim);
  font: inherit;
  font-size: 13px;
  cursor: default;
  transition: background .13s, color .13s;
}
nav button:hover { background: rgba(255,255,255,.06); color: #e9edf5 }
nav button.on {
  background: rgba(125,211,252,.14);
  color: var(--accent);
  box-shadow: 0 0 0 1px rgba(125,211,252,.22) inset;
}

.scroller {
  flex: 1;
  overflow-y: scroll;
  padding: 14px 16px 18px;
}
.scroller::-webkit-scrollbar { width: 8px }
.scroller::-webkit-scrollbar-track { background: transparent }
.scroller::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,.13);
  border-radius: 4px;
  border: 2px solid transparent;
  background-clip: padding-box;
}

h2 {
  margin: 16px 6px 8px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: .18em;
  text-transform: uppercase;
  color: rgba(233,237,245,.3);
}
h2:first-child { margin-top: 2px }

.ore {
  display: flex;
  align-items: center;
  padding: 8px 10px;
  border-radius: 11px;
  transition: background .13s, transform .13s;
}
.ore:hover { background: rgba(255,255,255,.07); transform: translateX(3px) }
.ore:active { background: rgba(255,255,255,.12) }
.ore .icon { --icon: 30px; margin-right: 13px }
.ore .name { flex: 1; text-transform: capitalize }
.ore .count {
  font-size: 12px;
  color: var(--dim);
  font-variant-numeric: tabular-nums;
}

.row {
  display: flex;
  align-items: center;
  padding: 13px 10px;
  border-bottom: 1px solid var(--line);
}
.row:last-child { border-bottom: 0 }
.row .label { flex: 1 }
.row .hint {
  display: block;
  font-size: 12px;
  color: var(--dim);
  margin-top: 2px;
}

.toggle {
  width: 42px;
  height: 24px;
  border-radius: 12px;
  background: rgba(255,255,255,.10);
  box-shadow: 0 0 0 1px rgba(255,255,255,.07) inset;
  position: relative;
  transition: background .15s;
}
.toggle.on { background: rgba(125,211,252,.55) }
.toggle i {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 18px;
  height: 18px;
  border-radius: 9px;
  background: #f4f7ff;
  box-shadow: 0 1px 3px rgba(0,0,0,.4);
  transition: transform .15s;
}
.toggle.on i { transform: translateX(18px) }

.steps { display: flex }
.steps span + span { margin-left: 6px }
.steps span {
  min-width: 42px;
  padding: 6px 0;
  text-align: center;
  border-radius: 8px;
  font-size: 13px;
  color: var(--dim);
  background: rgba(255,255,255,.05);
  transition: background .13s, color .13s;
}
.steps span:hover { background: rgba(255,255,255,.10); color: #e9edf5 }
.steps span.on {
  background: rgba(125,211,252,.16);
  color: var(--accent);
  box-shadow: 0 0 0 1px rgba(125,211,252,.24) inset;
}

footer {
  padding: 12px 20px;
  border-top: 1px solid var(--line);
  font-size: 12px;
  color: var(--dim);
}
footer kbd {
  display: inline-block;
  margin: 0 4px;
  padding: 2px 7px;
  border-radius: 5px;
  font: inherit;
  background: rgba(255,255,255,.08);
  box-shadow: 0 0 0 1px rgba(255,255,255,.08) inset;
}
`

const oreList = (bot: Bot) =>
  groups
    .map(
      group => `
      <h2>${group.label}</h2>
      ${group.ores
        .map(ore => {
          const held = bot.count(ore)

          return `
          <div class="ore" onclick="graft.send('mine','${ore}')">
            <i class="icon icon-${ore}"></i>
            <span class="name">${named(ore)}</span>
            <span class="count">${held > 0 ? held : ""}</span>
          </div>`
        })
        .join("")}`,
    )
    .join("")

const settings = () => `
  <div class="row">
    <span class="label">Manter no inventário
      <span class="hint">não descarta o que for minerado</span>
    </span>
    <div class="toggle ${manterInventario ? "on" : ""}" onclick="graft.send('toggle','guardar')"><i></i></div>
  </div>
  <div class="row">
    <span class="label">Voltar ao baú
      <span class="hint">guarda tudo quando o inventário enche</span>
    </span>
    <div class="toggle ${voltarAoBau ? "on" : ""}" onclick="graft.send('toggle','bau')"><i></i></div>
  </div>
  <div class="row">
    <span class="label">Raio de busca
      <span class="hint">blocos a partir de onde o bot está</span>
    </span>
    <div class="steps">
      ${[16, 32, 64, 128]
        .map(n => `<span class="${n === raio ? "on" : ""}" onclick="graft.send('raio','${n}')">${n}</span>`)
        .join("")}
    </div>
  </div>`

const panel = (bot: Bot) => `
  <div class="scrim"></div>
  <div class="shell">
    <header>
      <i class="icon icon-diamond_pickaxe" style="--icon:22px"></i>
      <h1>Minerar</h1>
      <span class="who">${bot.name}</span>
    </header>
    <nav>
      <button class="${tab === "minerar" ? "on" : ""}" onclick="graft.send('tab','minerar')">Minérios</button>
      <button class="${tab === "ajustes" ? "on" : ""}" onclick="graft.send('tab','ajustes')">Ajustes</button>
    </nav>
    <div class="scroller">${tab === "minerar" ? oreList(bot) : settings()}</div>
    <footer><kbd>B</kbd> fecha · <kbd>roda</kbd> rola a lista</footer>
  </div>`

export default plugin({
  name: "ore-panel",
  permissions: ["ui", "move", "dig", "chat", "inventory"],

  keys: {
    B: (bot, ui) =>
      ui.open({
        title: "Minerar",
        style,
        page: panel,
        on: {
          tab: name => {
            tab = name
          },
          raio: n => {
            raio = Number(n)
          },
          toggle: which => {
            if (which === "guardar") manterInventario = !manterInventario
            else voltarAoBau = !voltarAoBau
          },
          mine: (ore, bot) => {
            ui.dismiss()
            bot.say(`minerando ${named(ore)} num raio de ${raio}`)
            bot.pursue(mine(ore as Block))
          },
        },
      }),
  },
})
