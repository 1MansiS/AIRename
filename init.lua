local ok, rename = pcall(require, "ai_rename.rename")
if not ok then
  vim.notify("[AI Rename] Failed to load rename.lua", vim.log.levels.ERROR)
  return {}
end

pcall(vim.api.nvim_del_user_command, "AIRename")
vim.api.nvim_create_user_command("AIRename", function(opts)
  local provider = opts.args ~= "" and opts.args or nil
  local success, err = xpcall(function()
    rename.suggest_and_rename(provider)
  end, debug.traceback)
  if not success then
    vim.notify("[AI Rename] Runtime error:\n" .. err, vim.log.levels.ERROR)
  end
end, {
  nargs = "?",
  force = true,
  desc = "AI-assisted rename (file-scoped)",
})

return {}
