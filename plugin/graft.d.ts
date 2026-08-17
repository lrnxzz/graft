declare module "graft" {
  export interface Vec3 {
    x: number
    y: number
    z: number
  }

  export interface Vec2 {
    x: number
    z: number
  }

  export type Color = string

  export interface Area {
    radius?: number
    center?: Vec3
  }

  export type Block = "air" | "stone" | "dirt" | "diamond_ore"
  export type Item = "diamond_pickaxe" | "bread" | "torch"

  export interface Stack {
    item: Item
    count: number
    slot: number
  }

  export type Face = "up" | "down" | "north" | "south" | "east" | "west"

  export interface Target {
    at: Vec3
    face: Face
    block: Block
  }

  export interface Events {
    spawned: { at: Vec3 }
    arrived: { at: Vec3; reason: string }
    chat: { text: string }
    blockChanged: { at: Vec3; block: Block }
    health: { hp: number; food: number }
    disconnected: { reason: string }
  }

  export interface Intents {
    dig: { block: Vec3; tool: Item }
    place: { block: Vec3; item: Item }
    chat: { text: string }
    move: {}
  }

  export type Intent<T> = T & {
    cancel(reason?: string): void
    readonly cancelled: boolean
  }

  export interface Bot {
    readonly name: string
    readonly position: Vec3
    readonly health: number
    readonly food: number
    readonly onGround: boolean
    readonly held: Item
    readonly inventory: readonly Stack[]
    readonly looking: Target | null
    pursue(goal: Goal): Promise<void>
    abandon(): void
    goto(to: Vec3): Promise<Vec3>
    dig(block: Vec3): Promise<void>
    place(block: Vec3): Promise<void>
    hold(item: Item): Promise<void>
    say(text: string): Promise<void>
    look(at: Vec3): void
    blockAt(at: Vec3): Block
    count(item: Item): number
    on<K extends keyof Events>(event: K, handle: (e: Events[K]) => void): void
    before<K extends keyof Intents>(intent: K, handle: (e: Intent<Intents[K]>) => void): void
  }

  export interface Goal {
    readonly type: string
  }

  export function at(block: Vec3): Goal
  export function near(block: Vec3, radius: number): Goal
  export function mine(block: Block, where?: Area): Goal
  export function collect(item: Item, count?: number): Goal
  export function follow(player: string, opts?: { distance?: number }): Goal
  export function flee(from?: Vec3): Goal

  export function sequence(...goals: Goal[]): Goal
  export function repeat(goal: Goal, times?: number): Goal

  export function race(...goals: Goal[]): Goal
  export function until(goal: Goal, done: (bot: Bot) => boolean): Goal

  export interface Reaction {
    readonly type: "reaction"
  }

  export function when(
    condition: (bot: Bot) => boolean,
    act: (bot: Bot) => void | Promise<void>,
  ): Reaction

  export type ArgType = "string" | "number" | "block" | "item" | "player" | "position"

  type ArgValue<K extends ArgType> = K extends "number"
    ? number
    : K extends "position"
      ? Vec3
      : K extends "block"
        ? Block
        : K extends "item"
          ? Item
          : string

  export type ArgSpec = Record<string, ArgType | `${ArgType}?`>

  type Optional<K> = K extends `${string}?` ? true : false
  type Bare<K> = K extends `${infer B}?` ? B : K

  export type Args<S extends ArgSpec> = {
    [K in keyof S as Optional<S[K]> extends true ? never : K]: ArgValue<Bare<S[K]> & ArgType>
  } & {
    [K in keyof S as Optional<S[K]> extends true ? K : never]?: ArgValue<Bare<S[K]> & ArgType>
  }

  export interface Command {
    readonly type: "command"
  }

  export function command<S extends ArgSpec>(
    args: S,
    run: (bot: Bot, args: Args<S>) => void | Promise<void>,
    describe?: string,
  ): Command

  export type Anchor =
    | "top-left" | "top" | "top-right"
    | "left" | "center" | "right"
    | "bottom-left" | "bottom" | "bottom-right"

  export interface Element {
    readonly type: "element"
  }

  export interface PanelProps {
    anchor?: Anchor
    gap?: number
    padding?: number
    background?: Color
    children?: Element | Element[]
  }

  export interface TextProps {
    color?: Color
    scale?: number
    shadow?: boolean
    children?: string | number | Array<string | number>
  }

  export interface BarProps {
    value: number
    max: number
    color?: Color
    width?: number
  }

  export interface IconProps {
    item: Item | Block
    badge?: number | string
  }

  export interface ListProps {
    height: number
    gap?: number
    children?: Element | Element[]
  }

  export interface OptionProps {
    selected?: boolean
    onPick?: () => void
    children?: Element | Element[]
  }

  export function Panel(props: PanelProps): Element
  export function Row(props: PanelProps): Element
  export function Column(props: PanelProps): Element
  export function Text(props: TextProps): Element
  export function Bar(props: BarProps): Element
  export function Icon(props: IconProps): Element

  export function List(props: ListProps): Element

  export function Option(props: OptionProps): Element

  export function Raw(props: { draw: (canvas: Canvas) => void }): Element

  export interface Canvas {
    readonly width: number
    readonly height: number
    readonly cursor: Vec2
    fill(x: number, y: number, w: number, h: number, color: Color): void
    text(text: string, x: number, y: number, color?: Color, scale?: number): void
    icon(item: Item, x: number, y: number, size?: number): void
    measure(text: string, scale?: number): number
  }

  export interface Marker {
    readonly type: "marker"
  }

  export function highlight(block: Vec3, color?: Color): Marker
  export function box(from: Vec3, to: Vec3, color?: Color): Marker
  export function line(from: Vec3, to: Vec3, color?: Color): Marker
  export function beacon(at: Vec3, color?: Color): Marker

  interface Framed {
    title?: string
    onKey?: (key: Key, bot: Bot) => void
  }

  interface Built extends Framed {
    body: (bot: Bot) => Element
  }

  interface Written extends Framed {
    page: (bot: Bot) => string
    style?: string
    on?: Record<string, (...args: [...string[], Bot]) => void>
  }

  export type Menu = Built | Written

  export interface Ui {
    open(menu: Menu): void
    dismiss(): void
    readonly showing: boolean
    toast(text: string): void
  }

  export type Key =
    | "A" | "B" | "C" | "D" | "E" | "F" | "G" | "H" | "I" | "J" | "K" | "L" | "M"
    | "N" | "O" | "P" | "Q" | "R" | "S" | "T" | "U" | "V" | "W" | "X" | "Y" | "Z"
    | "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"
    | "Escape" | "Enter" | "Space" | "Tab"

  export type Permission = "move" | "dig" | "place" | "chat" | "inventory" | "ui"

  export interface Plugin {
    name: string
    version?: string
    describe?: string
    permissions?: Permission[]
    reactions?: Reaction[]
    commands?: Record<string, Command>
    hud?: (bot: Bot) => Element | null
    world?: (bot: Bot) => Marker[]
    keys?: Partial<Record<Key, (bot: Bot, ui: Ui) => void>>
    setup?: (bot: Bot, ui: Ui) => void | Promise<void>
    teardown?: () => void
  }

  export function plugin(spec: Plugin): Plugin

  export namespace JSX {
    interface ElementAttributesProperty {
      props: unknown
    }
    type Element = import("graft").Element
  }
}
