vim.lsp.enable({"gopls"})

local o = vim.o
local au = vim.api.nvim_create_autocmd
local map = vim.keymap.set


o.number = false
o.relativenumber = false
o.cursorline = true
o.wrap = true
o.scrolloff = 15
o.sidescrolloff = 15

o.tabstop = 4
o.shiftwidth = 4
o.softtabstop = 4
o.smartindent = true

o.ignorecase = true
o.smartcase = true
o.hlsearch = true
o.incsearch = true

o.showmatch = true
o.showmode = false
o.winblend = 0
o.concealcursor = ""
o.lazyredraw = true

o.confirm = true
o.swapfile = false
o.backup = false
o.writebackup = false
o.undofile = true
o.errorbells = false
o.clipboard = "unnamedplus"

o.splitbelow = true
o.splitright = true

vim.g.mapleader = " "
vim.g.localmapleader = " "

map({ "n", "v", "o" }, "<a-k>", "gg", { desc = "File:Top" })
map({ "n", "v", "o" }, "<a-j>", "G",  { desc = "File:Bottom" })
map({ "n", "v", "o" }, "<a-h>", "0",  { desc = "Line:End" })
map({ "n", "v", "o" }, "<a-l>", "$",  { desc = "Line:Beginning" })

map("n", "<c-k>", "<c-w>k", { desc = "Window:Up" })
map("n", "<c-j>", "<c-w>j", { desc = "Window:Bottom" })
map("n", "<c-h>", "<c-w>h", { desc = "Window:End" })
map("n", "<c-l>", "<c-w>l", { desc = "Window:Beginning" })

map("n", "<esc>", vim.cmd.nohlsearch, { desc = "Hide Search Highlights" })

map("n", "<leader>w", "<c-w>", { desc = "Windows" })

map("n", "<leader>bd", vim.cmd.bdelete, { desc = "Buffer:Delete" })
map("n", "<s-h>", vim.cmd.bprev, { desc = "Buffer:Prev" })
map("n", "<s-l>", vim.cmd.bnext, { desc = "Buffer:Next" })

au("VimResized", {
	callback = function()
		vim.cmd("wincmd =")
	end
})

au("TextYankPost", {
	callback = function()
		vim.hl.on_yank()
	end
})

vim.pack.add({
	"https://github.com/nvim-mini/mini.nvim",
	"https://github.com/neovim/nvim-lspconfig",
})

require("mini.icons").setup()
MiniIcons.mock_nvim_web_devicons()

require("mini.ai").setup()
require("mini.align").setup()
require("mini.comment").setup()
require("mini.pairs").setup()
require("mini.splitjoin").setup()
require("mini.surround").setup()
require("mini.bracketed").setup()
require("mini.cmdline").setup()
require("mini.files").setup()
map("n", "<leader>e", MiniFiles.open, { desc = "Mini Files" })
require("mini.jump").setup()
require("mini.jump2d").setup({
	view = {
		dim = true
	}
})
require("mini.hipatterns").setup()
require("mini.indentscope").setup({
	symbol = "│"	
})
require("mini.notify").setup()
require("mini.starter").setup()
require("mini.statusline").setup()
require("mini.tabline").setup()
require("mini.trailspace").setup()
require("mini.git").setup()
require("mini.diff").setup({
	view = {
	style = "sign",
		signs = { add = "│", change = "│", delete = "│" }
	}
})
require("mini.extra").setup()

require("mini.completion").setup({
	lsp_completion = {
		source_func = "omnifunc"
	}
})

require('mini.hipatterns').setup()
MiniHipatterns.setup({
  highlighters = {
    fixme = { pattern = '%f[%w]()FIXME()%f[%W]', group = 'MiniHipatternsFixme' },
    hack  = { pattern = '%f[%w]()HACK()%f[%W]',  group = 'MiniHipatternsHack'  },
    todo  = { pattern = '%f[%w]()TODO()%f[%W]',  group = 'MiniHipatternsTodo'  },
    note  = { pattern = '%f[%w]()NOTE()%f[%W]',  group = 'MiniHipatternsNote'  },
    hex_color = MiniHipatterns.gen_highlighter.hex_color(),
  },
})

require('mini.clue').setup()
MiniClue.setup({
  triggers = {
    { mode = { 'n', 'x' }, keys = '<Leader>' },
    { mode = 'n', keys = '[' },
    { mode = 'n', keys = ']' },
    { mode = 'i', keys = '<C-x>' },
    { mode = { 'n', 'x' }, keys = 'g' },
    { mode = { 'n', 'x' }, keys = "'" },
    { mode = { 'n', 'x' }, keys = '`' },
    { mode = { 'n', 'x' }, keys = '"' },
    { mode = { 'i', 'c' }, keys = '<C-r>' },
    { mode = 'n', keys = '<C-w>' },
    { mode = { 'n', 'x' }, keys = 'z' },
  },

  clues = {
    MiniClue.gen_clues.square_brackets(),
    MiniClue.gen_clues.builtin_completion(),
    MiniClue.gen_clues.g(),
    MiniClue.gen_clues.marks(),
    MiniClue.gen_clues.registers(),
    MiniClue.gen_clues.windows(),
    MiniClue.gen_clues.z(),
  },

  window = {
	  delay = 0
  }
})

require("mini.pick").setup()
vim.ui.select = MiniPick.ui_select

map("n", "<leader>s", "<Nop>", { desc = "Search" })
map("n", "<leader>f", MiniPick.builtin.files, { desc = "Search Files" })
map("n", "<leader>g", MiniPick.builtin.grep_live, { desc = "Grep" })
map("n", "<leader>/", MiniExtra.pickers.buf_lines, { desc = "Lines" })
map("n", "<leader>sr", MiniPick.builtin.resume, { desc = "Resume" })
map("n", "<leader>sk", MiniExtra.pickers.keymaps, { desc = "Keymaps" })

require("zenkai")
