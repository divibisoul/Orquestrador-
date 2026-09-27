import unittest

from graph import aggregate, build_graph, decompose


class GraphTests(unittest.TestCase):
    def test_graph_compiles(self):
        self.assertIsNotNone(build_graph())

    def test_decompose_normalizes_tasks(self):
        result = decompose({"tasks": [{"id": "a", "prompt": "hello"}]})
        self.assertEqual(
            result["tasks"],
            [{"id": "a", "prompt": "hello", "model": "soul-auto"}],
        )
        self.assertEqual(result["routed"], [])

    def test_aggregate_preserves_routed_results(self):
        routed = [{"id": "a", "status": "routed", "text": "ok", "metadata": {}}]
        result = aggregate({"tasks": [], "routed": routed})
        self.assertEqual(result["aggregate"]["count"], 1)
        self.assertEqual(result["aggregate"]["results"], routed)


if __name__ == "__main__":
    unittest.main()
