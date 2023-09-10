function* numericalIdGenerator(seed = Date.now()) {
    while (true) yield seed++;
  }
  
  const generateNumericalId = numericalIdGenerator();
  
  export const genId = () => generateNumericalId.next().value;