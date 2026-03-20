// app/theme/defaults.ts
const SURFACE = {
  rounded: 'lg',
  elevation: 0,
} as const

const BUTTON = {
  rounded: 'lg',
  elevation: 0,
} as const

const FIELD = {
  rounded: 'lg',
  variant: 'outlined',
} as const

const CHIPLIKE = {
  rounded: 'lg',
} as const

const defaults = {
  global: {
    rounded: 'lg',
    elevation: 0,
  },

  // Core surfaces / containers
  VAlert: SURFACE,
  VAppBar: {
    rounded: 0,
    elevation: 0,
  },
  VBanner: SURFACE,
  VBottomNavigation: SURFACE,
  VBottomSheet: {
    rounded: 'lg',
    elevation: 0,
  },
  VCard: SURFACE,
  VCarousel: SURFACE,
  VCode: CHIPLIKE,
  VDialog: {
    rounded: 'lg',
  },
  VExpansionPanel: SURFACE,
  VExpansionPanels: {
    rounded: 'lg',
  },
  VKbd: CHIPLIKE,
  VList: {
    rounded: 'lg',
  },
  VListItem: {
    rounded: 'lg',
  },
  VMenu: {
    rounded: 'lg',
    elevation: 0,
  },
  VNavigationDrawer: {
    rounded: 0,
    elevation: 0,
  },
  VOverlay: {
    rounded: 'lg',
  },
  VSheet: SURFACE,
  VSnackbar: SURFACE,
  VTable: {
    rounded: 'lg',
  },
  VToolbar: SURFACE,
  VWindow: {
    rounded: 'lg',
  },

  // Buttons / clickable controls
  VBtn: BUTTON,
  VBtnGroup: {
    rounded: 'lg',
  },
  VBtnToggle: {
    rounded: 'lg',
  },
  VFab: BUTTON,
  VIconBtn: BUTTON,
  VPagination: {
    rounded: 'lg',
  },
  VChip: CHIPLIKE,

  // Form field foundation
  VField: FIELD,
  VInput: {
    density: 'comfortable',
  },

  // Text-like inputs
  VTextField: FIELD,
  VTextarea: FIELD,
  VAutocomplete: FIELD,
  VCombobox: FIELD,
  VFileInput: FIELD,
  VSelect: { ...FIELD, flat: true },

  // Pickers / structured inputs
  VColorPicker: SURFACE,
  VDatePicker: SURFACE,

  // Data display
  VDataTable: {
    rounded: 'lg',
    elevation: 0,
  },
  VDataIterator: {
    rounded: 'lg',
  },

  // Media / misc
  VAvatar: {
    rounded: 'lg',
  },
  VImg: {
    rounded: 'lg',
  },
  VSkeletonLoader: {
    rounded: 'lg',
    elevation: 0,
  },
} as const

export default defaults