local M = {}

local ts = vim.treesitter

local CLI = vim.fn.expand("~/.config/nvim/lua/ai_rename/go/ai_rename_bin")

-- ----------------------------
-- Treesitter helpers
-- ----------------------------
local function get_node_under_cursor()
  return ts.get_node()
end

local function get_node_text(node)
  local bufnr = vim.api.nvim_get_current_buf()
  local text = ts.get_node_text(node, bufnr)
  if type(text) == "table" then
    return table.concat(text, "")
  end
  return text
end

-- ----------------------------
-- CLI invocation
-- ----------------------------

local function run_cli(filepath, symbol, provider)
  local cmd = {
    CLI,
    "-llm", provider,
    filepath,
    symbol,
  }

  local output = vim.fn.system(cmd)
  if vim.v.shell_error ~= 0 then
    return nil, output
  end

  local json_str = output:match("(%b{})")
  if not json_str then
    return nil, "No JSON found in CLI output:\n" .. output
  end

  local ok, result = pcall(vim.json.decode, json_str)
  if not ok then
    return nil, "JSON decode error: " .. result
  end
  return result
end

-- ----------------------------
-- Entry point
-- ----------------------------

function M.suggest_and_rename(provider)
  provider = provider or "claude"
  local node = get_node_under_cursor()
  if not node then return end

  local name = get_node_text(node)
  if not name or name == "" then return end

  local win = vim.api.nvim_get_current_win()
  local pos = vim.api.nvim_win_get_cursor(win)
  local filepath = vim.api.nvim_buf_get_name(0)
  local symbol = pos[1] .. ":" .. pos[2]

  local result, _ = run_cli(filepath, symbol, provider)
  if not result then return end

  -- Filter out suggestions that match the current name
  local suggestions = vim.tbl_filter(function(s)
    return s.name ~= name
  end, result.suggestions or {})

  if #suggestions == 0 then return end

  vim.ui.select(suggestions, {
    prompt = "Rename `" .. name .. "`",
    format_item = function(item)
      return item.name .. " - " .. (item.reason or "")
    end,
  }, function(choice)
    if not choice then return end

    local bufnr = vim.api.nvim_win_get_buf(win)
    local clients = vim.lsp.get_clients({ bufnr = bufnr })
    if #clients == 0 then return end

    -- Restore cursor on the original window without switching focus.
    -- nvim_set_current_win triggers WinEnter autocmds that can move the cursor
    -- to a different identifier before the rename fires.
    vim.api.nvim_win_set_cursor(win, pos)

    local enc = clients[1].offset_encoding or "utf-16"
    local params = vim.lsp.util.make_position_params(win, enc)
    params.newName = choice.name

    vim.lsp.buf_request(bufnr, "textDocument/rename", params, function(err, result, ctx)
      if err or not result then return end
      local client = vim.lsp.get_client_by_id(ctx.client_id)
      local apply_enc = client and client.offset_encoding or enc
      local orig_notify = vim.notify
      vim.notify = function() end
      local save_shortmess = vim.o.shortmess
      vim.opt.shortmess:append("F")
      vim.lsp.util.apply_workspace_edit(result, apply_enc)
      vim.o.shortmess = save_shortmess
      vim.notify = orig_notify
      vim.cmd("echo ''")
    end)
  end)
end

return M
