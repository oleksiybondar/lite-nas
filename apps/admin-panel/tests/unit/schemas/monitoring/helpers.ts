/**
 * Asserts that one monitoring snapshot parser normalizes a missing record list and parses data.
 */
export const expectMonitoringSnapshotParser = <TResult>({
  emptyEnvelope,
  expectedEmptyResult,
  expectedParsedResult,
  parser,
  populatedEnvelope,
}: {
  emptyEnvelope: unknown;
  expectedEmptyResult: TResult;
  expectedParsedResult: TResult;
  parser: (value: unknown) => TResult;
  populatedEnvelope: unknown;
}): void => {
  expect(parser(emptyEnvelope)).toEqual(expectedEmptyResult);
  expect(parser(populatedEnvelope)).toEqual(expectedParsedResult);
};
