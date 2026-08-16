import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
  resolveScaledContentMinHeight,
  resolveScaledControlMinHeight,
} from "../../core/dist/index.js";

const fixture = JSON.parse(
  await readFile(
    new URL("../testdata/typography-scaling.fixture.json", import.meta.url),
  ),
);

test("keeps control and feedback content reachable at 100, 160, and 235 percent", () => {
  const controlHeights = fixture.scales.map((fontScale) =>
    resolveScaledControlMinHeight({
      baseHeight: 52,
      fontScale,
      lineHeight: 24,
      maximumFontScale: 2.35,
      verticalPadding: 8,
    }),
  );
  const feedbackHeights = fixture.scales.map((fontScale) =>
    resolveScaledContentMinHeight({
      fontScale,
      gap: 8,
      lineHeights: [24, 24],
      maximumFontScale: 2.35,
      verticalPadding: 8,
    }),
  );

  assert.deepEqual(controlHeights, fixture.singleLineControlMinHeights);
  assert.deepEqual(feedbackHeights, fixture.twoLineFeedbackMinHeights);
});
