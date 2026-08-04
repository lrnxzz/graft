import { plugin } from "gocraft"

export const seen: string[] = []

export default plugin({
  name: "listener",
  permissions: ["chat"],

  setup: bot => {
    bot.on("chat", e => seen.push("chat:" + e.text))
    bot.on("nonsense", () => seen.push("nonsense"))

    bot.before("dig", e => e.cancel("nothing is mined here"))
  },
})
