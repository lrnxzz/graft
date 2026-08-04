import { plugin, command } from "gocraft"

export default plugin({
  name: "beta",
  permissions: ["chat"],

  commands: {
    mine: command({}, bot => bot.say("beta wins?")),
    there: command({}, bot => bot.say("beta")),
  },

  keys: {
    M: bot => bot.say("beta key"),
  },
})
