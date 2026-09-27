from graph import aggregate, build_graph, decompose


def test_graph_compiles():
    assert build_graph() is not None


def test_decompose_normalizes_tasks():
    result = decompose({"tasks": [{"id": "a", "prompt": "hello"}]})
    assert result["tasks"] == [{"id": "a", "prompt": "hello", "model": "soul-auto"}]
    assert result["routed"] == []


def test_aggregate_preserves_routed_results():
    routed = [{"id": "a", "status": "routed", "text": "ok", "metadata": {}}]
    result = aggregate({"tasks": [], "routed": routed})
    assert result["aggregate"]["count"] == 1
    assert result["aggregate"]["results"] == routed
