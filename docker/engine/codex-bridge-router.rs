// CommonRoom's dedicated build has no native/MCP/hosted/Code Mode tool routes.
// A source fork CAN restore them, BUT exposes engine-held provider credentials.
fn commonroom_bridge_router(
    dynamic_tools: &[DynamicToolSpec],
    policy: Arc<codex_extension_api::ToolPolicy>,
) -> CodexResult<ToolRouter> {
    let [DynamicToolSpec::Function(tool)] = dynamic_tools else {
        return Err(CodexErrorDetails::InvalidRequest("CommonRoom requires exactly one bridge tool".into()).into());
    };
    if tool.name != "commonroom_tool" || tool.defer_loading {
        return Err(CodexErrorDetails::InvalidRequest("CommonRoom bridge tool unavailable".into()).into());
    }
    let Some(handler) = DynamicToolHandler::new(tool) else {
        return Err(CodexErrorDetails::InvalidRequest("CommonRoom bridge schema invalid".into()).into());
    };
    let mut restricted = (*policy).clone();
    let name = codex_tools::ToolName::plain("commonroom_tool");
    restricted.allowed_tools = Some(if policy.allows(&name) { vec![name] } else { Vec::new() });
    let mut registry = ToolRegistry::with_tool_policy(Arc::new(restricted));
    registry.register_external(Arc::new(handler));
    let specs = registry.entries().map(|entry| entry.runtime.spec()).collect();
    Ok(ToolRouter::from_parts(registry, specs, ToolMode::Direct, BTreeMap::new(), None, &[]))
}

#[cfg(test)]
mod commonroom_bridge_tests {
    use super::*;
    use codex_protocol::dynamic_tools::DynamicToolFunctionSpec;
    fn bridge(name: &str) -> DynamicToolSpec {
        DynamicToolSpec::Function(DynamicToolFunctionSpec {
            name: name.into(), description: "Scoped CommonRoom broker".into(),
            input_schema: serde_json::json!({"type":"object","properties":{},"additionalProperties":false}),
            defer_loading: false,
        })
    }
    #[test]
    fn commonroom_bridge_has_only_direct_callback_no_native_or_code_mode() {
        let router = commonroom_bridge_router(&[bridge("commonroom_tool")], Arc::new(Default::default())).unwrap();
        let specs = router.model_visible_specs();
        assert_eq!(specs.len(), 1);
        assert_eq!(specs[0].name(), "commonroom_tool");
        assert_eq!(router.tool_mode(), ToolMode::Direct);
        assert!(!router.requires_code_mode_worker());
        assert!(!router.can_manage_children());
        assert!(!router.has_terminal_controls());
    }
    #[test]
    fn commonroom_bridge_refuses_other_tools_and_broad_restored_sessions() {
        for tools in [vec![],vec![bridge("exec_command")],vec![bridge("apply_patch")],vec![bridge("commonroom_tool"),bridge("exec_command")]] {
            assert!(commonroom_bridge_router(&tools, Arc::new(Default::default())).is_err());
        }
        let mut deferred = bridge("commonroom_tool");
        let DynamicToolSpec::Function(ref mut tool) = deferred else { unreachable!() };
        tool.defer_loading = true;
        assert!(commonroom_bridge_router(&[deferred],Arc::new(Default::default())).is_err());
    }
    #[test]
    fn commonroom_bridge_cannot_relax_host_tool_ceiling() {
        let policy = codex_extension_api::ToolPolicy {allowed_tools:Some(vec![]),..Default::default()};
        let router = commonroom_bridge_router(&[bridge("commonroom_tool")],Arc::new(policy)).unwrap();
        assert!(router.model_visible_specs().is_empty());
    }
}
