# @ansldes/core

Compiler가 생성한 resolved bundle을 framework와 무관하게 소비하는 runtime이다.

```ts
const designSystem = createDesignSystem(bundle, {
  theme: "light",
});

designSystem.color("semantic", "text.primary");
designSystem.typography("body");
designSystem.component("button");
designSystem.icon("arrow-right");
designSystem.iconUsage("navigation.next");
designSystem.motionTransition("control.press", reduceMotion);
designSystem.elevation("raised");
```

Runtime은 제품 token 이름을 미리 알지 않는다. 제품이 definition에서 제공한 이름과 component recipe만
조회한다.

`resolveScaledControlMinHeight`와 `resolveScaledContentMinHeight`는 framework가 달라도 같은 large-text
계산을 재사용하게 한다. `resolveToggleThumbTravel`, `mapSelectionThroughFormat`과
`mapDigitSelectionThroughFormat`은 제품 값 없이 interaction geometry와 formatter selection을 계산한다.
유효하지 않은 theme, token, icon 또는 component는 즉시 실패한다. React Native의 font-weight 표현과
SVG rendering은 framework-neutral core가 아니라 platform adapter가 소유한다.
