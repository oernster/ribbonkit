// ribbonkit's half of the page: what every ribbon's page does about its window, whatever it shows.
// An application imports it from '@oernster/ribbonkit'; test helpers from '@oernster/ribbonkit/testing'.

export { Band } from './Band'
export { backgroundReporter, rgbOf, rootId, swatchId, type Rgb } from './background'
export { About, Licence, Update } from './Help'
export { connect, on, startDrag, windowCalls, type Call, type Refused, type WindowBridge, type WindowCalls } from './bridge'
export { showsTheMenu, useDrag, type Distance } from './drag'
export { MenuGroup, MenuToggle } from './MenuChoices'
export { opacityProperty, percentOfWhole, showOpacity } from './opacity'
export { OpacitySlider } from './OpacitySlider'
export { naturalHeight, usePanelFit } from './panelFit'
export { watchPixelRatio } from './pixelRatio'
export { PullOut, type PullOutWords } from './PullOut'
export { ScaleGrip } from './ScaleGrip'
export { scrollbarThickness } from './scrollbar'
export { useShell, type Drawn, type Panel, type ShellOptions, type View } from './shell'
export type { AboutFacts, Box, Credit, MenuChoice, UpdateStatus } from './wire'
