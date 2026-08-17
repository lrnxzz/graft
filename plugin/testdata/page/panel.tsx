import { plugin } from "graft"

export const picked: string[] = []

export default plugin({
  name: "panel",
  permissions: ["ui", "chat"],

  keys: {
    M: (bot, ui) =>
      ui.open({
        title: "Ores",
        style: `.ore { padding: 8px; border-radius: 6px }
                .ore:hover { background: #ffffff22 }`,
        page: bot => `
          <div class="list">
            <p>${bot.name}</p>
            <div class="ore" onclick="graft.send('pick', 'stone')">stone</div>
            <div class="ore" onclick="graft.send('pick', 'iron_ore')">iron</div>
          </div>`,
        on: {
          pick: (block, bot) => {
            picked.push(block)
            bot.say("mining " + block)
          },
        },
      }),
  },
})
