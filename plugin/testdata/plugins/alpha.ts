import { plugin, command } from "graft"

export default plugin({
  name: "alpha",
  permissions: ["move", "chat"],

  commands: {
    mine: command({ ore: "block" }, (bot, { ore }) => bot.say(ore)),
    here: command({}, bot => bot.say("alpha")),
  },

  keys: {
    M: bot => bot.abandon(),
  },
})
