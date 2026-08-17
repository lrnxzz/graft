import {
  plugin, command, when,
  mine, collect, follow, flee, sequence, repeat, within,
  highlight, beacon,
  Panel, Row, Text, Bar, Icon, List, Option, Raw,
  type Block, type Bot, type Ui,
} from "graft"

const mineable: Block[] = [
  "coal_ore", "deepslate_coal_ore",
  "copper_ore", "deepslate_copper_ore",
  "iron_ore", "deepslate_iron_ore",
  "gold_ore", "deepslate_gold_ore",
  "redstone_ore", "deepslate_redstone_ore",
  "lapis_ore", "deepslate_lapis_ore",
  "diamond_ore", "deepslate_diamond_ore",
  "emerald_ore", "deepslate_emerald_ore",
  "nether_quartz_ore", "nether_gold_ore", "ancient_debris",
]

const picked = new Set<Block>(["diamond_ore", "deepslate_diamond_ore"])

function toggle(ore: Block) {
  if (picked.has(ore)) {
    picked.delete(ore)

    return
  }

  picked.add(ore)
}

function pursuePicked(bot: Bot) {
  if (picked.size === 0) {
    bot.say("nenhum minério selecionado")

    return
  }

  const rounds = [...picked].map(ore => mine(ore, within(64)))
  bot.pursue(repeat(sequence(...rounds)))
}

const orePicker = {
  title: "o que minerar",

  body: (bot: Bot) => (
    <Panel anchor="center" gap={6} padding={12} background="#111d">
      <Text scale={2}>o que minerar</Text>

      <List height={180} gap={2}>
        {mineable.map(ore => (
          <Option selected={picked.has(ore)} onPick={() => toggle(ore)}>
            <Row gap={4}>
              <Icon item={ore} />
              <Text color={picked.has(ore) ? "#0f8" : "#ccc"}>{ore}</Text>
            </Row>
          </Option>
        ))}
      </List>

      <Row gap={6}>
        <Text color="#888">{picked.size} selecionado(s)</Text>
        <Text color="#888">enter para começar · esc para fechar</Text>
      </Row>

      <Bar value={bot.health} max={20} color="#e33" width={160} />
    </Panel>
  ),

  onKey: (key, bot) => {
    if (key === "Enter") {
      pursuePicked(bot)
    }
  },
}

export default plugin({
  name: "auto-miner",
  version: "1.0.0",
  describe: "Mines an ore on command and looks after itself while doing it",

  permissions: ["move", "dig", "inventory", "chat", "ui"],

  reactions: [
    when(bot => bot.health < 6, bot => bot.pursue(flee())),
    when(bot => bot.food < 14, bot => bot.hold("bread")),
    when(bot => bot.count("diamond") >= 64, bot => bot.say("inventário cheio")),
  ],

  commands: {
    mine: command(
      { ore: "block", radius: "number?" },
      (bot, { ore, radius = 64 }) => bot.pursue(repeat(mine(ore, within(radius)))),
      "mina um minério até o inventário encher",
    ),

    fetch: command(
      { item: "item", count: "number?" },
      (bot, { item, count = 1 }) => bot.pursue(collect(item, count)),
    ),

    come: command(
      { player: "player" },
      (bot, { player }) => bot.pursue(follow(player, { distance: 3 })),
    ),

    stop: command({}, bot => bot.abandon()),
  },

  hud: bot => (
    <Panel anchor="top-left" gap={4} padding={8} background="#000a">
      <Text scale={2}>{bot.name}</Text>

      <Bar value={bot.health} max={20} color="#e33" />
      <Bar value={bot.food} max={20} color="#b83" />

      <Row gap={2}>
        {[...picked].map(ore => (
          <Icon item={ore} />
        ))}
      </Row>

      <Raw draw={c => c.text(`${c.cursor.x | 0},${c.cursor.z | 0}`, 0, 0, "#888")} />
    </Panel>
  ),

  world: bot => {
    const target = bot.looking
    if (!target) {
      return []
    }

    return [
      highlight(target.at, "#fc0"),
      beacon(bot.position, "#0f8"),
    ]
  },

  keys: {
    M: (bot, ui) => ui.open(orePicker),
    K: bot => bot.abandon(),
  },

  setup: async (bot: Bot, ui: Ui) => {
    bot.before("dig", e => {
      if (e.block.y < -50) {
        e.cancel("fundo do mundo é proibido")
      }
    })

    bot.on("arrived", ({ at }) => ui.toast(`cheguei em ${at.x},${at.y},${at.z}`))

    await bot.pursue(collect("torch", 32))
  },
})
