-- Neovim for the TOT lab VM.
--
-- The VM has no network, so nothing may be fetched at startup: lazy.nvim is
-- cloned and every plugin is installed and compiled during the image build
-- (`nvim --headless "+Lazy! sync"`), and the update checker is off so a student
-- opening a file never waits on a request that cannot succeed.
--
-- The plugin set is deliberately small. This is an editor a student meets for a
-- few minutes inside a lesson, not a workstation: they need to see the YAML they
-- are writing, know where they are, and be able to find a command they half
-- remember. Anything beyond that is weight they did not ask for.
--
-- No treesitter. Its parsers have to be compiled into the image, and getting
-- that to happen reliably in a headless build cost four attempts for a gain
-- Neovim's built-in syntax already covers at the scale of a lesson: YAML, bash,
-- Dockerfile, JSON and Markdown all highlight fine without it.

vim.g.mapleader = " "
vim.g.maplocalleader = " "

local o = vim.opt
o.number = true
o.relativenumber = false
o.mouse = "a"
o.clipboard = ""            -- no X server in the VM; keep the system register out of it
o.expandtab = true          -- YAML dies on tabs, and half the labs are YAML
o.shiftwidth = 2
o.tabstop = 2
o.softtabstop = 2
o.smartindent = true
o.termguicolors = true
o.signcolumn = "yes"
o.cursorline = true
o.wrap = false
o.scrolloff = 4
o.ignorecase = true
o.smartcase = true
o.undofile = true
o.updatetime = 250
o.timeoutlen = 400
o.splitright = true
o.splitbelow = true
o.list = true
o.listchars = { tab = "→ ", trail = "·", nbsp = "␣" }

-- Trailing whitespace and tabs are invisible bugs in a manifest; make both show.
vim.api.nvim_create_autocmd("FileType", {
  pattern = { "yaml", "yml", "json", "helm" },
  callback = function()
    vim.opt_local.shiftwidth = 2
    vim.opt_local.tabstop = 2
  end,
})

-- A few keys a first-time user can guess. Everything else is stock Neovim, so
-- what the lesson teaches about vim still applies.
local map = vim.keymap.set
map("n", "<leader>w", "<cmd>w<cr>", { desc = "Сохранить" })
map("n", "<leader>q", "<cmd>q<cr>", { desc = "Выйти" })
map("n", "<leader>e", "<cmd>NvimTreeToggle<cr>", { desc = "Дерево файлов" })
map("n", "<Esc>", "<cmd>nohlsearch<cr>", { desc = "Снять подсветку поиска" })

local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
vim.opt.rtp:prepend(lazypath)

require("lazy").setup({
  {
    "folke/tokyonight.nvim",
    lazy = false,
    priority = 1000,
    config = function()
      require("tokyonight").setup({ style = "night", transparent = false })
      vim.cmd.colorscheme("tokyonight")
    end,
  },

  {
    "nvim-lualine/lualine.nvim",
    dependencies = { "nvim-tree/nvim-web-devicons" },
    event = "VeryLazy",
    config = function()
      require("lualine").setup({
        options = { theme = "tokyonight", icons_enabled = true, globalstatus = true },
      })
    end,
  },

  {
    "nvim-tree/nvim-tree.lua",
    dependencies = { "nvim-tree/nvim-web-devicons" },
    cmd = "NvimTreeToggle",
    config = function()
      require("nvim-tree").setup({ view = { width = 30 } })
    end,
  },

  {
    "lewis6991/gitsigns.nvim",
    event = "VeryLazy",
    config = function()
      require("gitsigns").setup()
    end,
  },

  {
    "lukas-reineke/indent-blankline.nvim",
    main = "ibl",
    event = "VeryLazy",
    opts = {},
  },

  -- Shows what <leader> can do. The one plugin here that exists because the
  -- reader is learning rather than working.
  {
    "folke/which-key.nvim",
    event = "VeryLazy",
    opts = {},
  },
}, {
  -- Offline guarantees: never check for updates, never report them.
  checker = { enabled = false },
  change_detection = { enabled = false, notify = false },
  -- install.missing НЕ выключаем: этот же флаг блокирует установку во время
  -- сборки образа, и первая попытка оставила один lazy.nvim без единого плагина.
  -- После успешной сборки недостающих плагинов не остаётся, а проверку
  -- обновлений выключает checker выше — в офлайн-VM никто никуда не ходит.
  ui = { border = "rounded" },
  -- Свой путь для lazy-lock.json: по умолчанию он пишется рядом с конфигом, а в
  -- собранном образе тот каталог только для чтения — запись падала assert'ом и
  -- роняла сборку уже ПОСЛЕ того, как плагины успешно встали.
  lockfile = vim.fn.stdpath("state") .. "/lazy-lock.json",
})
