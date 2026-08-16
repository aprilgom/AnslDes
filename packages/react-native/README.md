# @ansldes/react-native

제품 compiled bundle이 제공하는 component recipe를 렌더링하는 React Native runtime이다.

```tsx
<DesignSystemProvider runtime={runtime}>
  <Button label="Continue" onPress={submit} size="medium" variant="primary" />
</DesignSystemProvider>
```

제공하는 기본 anatomy:

- `DesignText`: semantic typography role과 font scaling
- `Button`: 상태, focus border, loading과 large-text height
- `TextField`: label/input/message 순서와 focus/error 상태
- `SelectionControl`: checkbox/radio role, checked state와 indicator
- `ListItem`: 정적/interactive row, copy와 accessory 경계
- `Feedback`: polite/assertive live-region과 action slot

각 component는 제품 color/radius/token 이름을 하드코딩하지 않는다. 제품 definition의 `slots`, `variants`,
`sizes`가 제공한 resolved 값만 소비한다. `react`와 `react-native`는 peer dependency이며 runtime bundle에
포함하지 않는다.

저장소 검사는 좁은 compile-time React Native port type을 사용해 upstream Metro를 설치하지 않는다. 실제
React Native 0.85.3 타입에 대한 별도 peer 검증도 수행했으며, 100%·160%·235% fixture가 control과 multi-line
feedback의 최소 높이를 고정한다.
