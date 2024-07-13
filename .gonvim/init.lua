return {
    lsp_cfg = {
        settings = {
            gopls = {
                buildFlags = {"-tags=redis,rabbitmq,kafka"}
            }
        }
    }
}
