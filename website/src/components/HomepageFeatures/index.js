import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'AI Metatadata Store',
    Svg: require('@site/static/img/undraw_docusaurus_mountain.svg').default,
    description: (
      <>
      Log, Search and Compare metadata from experiments, models and checkpoints.
      Support for logging metrics and events from various services involved in training
      and deployment of models. 
      </>
    ),
  },
  {
    title: 'Model Registry',
    Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
    description: (
      <>
      Track and store trained models or upload them to a built in artifact server. Automatically
      apply recipes to transform models to representations understood by inference platforms
      and runtimes.
      </>
    ),
  },
  {
    title: 'Continuous Model Evaluation',
    Svg: require('@site/static/img/undraw_docusaurus_react.svg').default,
    description: (
      <>
         Measure performance metrics such as throughput and accuracy of models.
         Write deployment checks based on metrics to automatically tag models to promote to production.
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
