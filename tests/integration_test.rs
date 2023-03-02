use time::{PrimitiveDateTime, OffsetDateTime};

mod setup;

use modelbox::repository;

fn now() -> PrimitiveDateTime {
    let n = OffsetDateTime::now_utc();
    PrimitiveDateTime::new(n.date(), n.time())
}

#[tokio::test]
async fn test_create_example() {
    let db = setup::create_db().await.unwrap();
    let repository = repository::Repository::new_with_db(db);
    let experiment = entity::experiments::Model {
        id: "abcd".into(),
        name: "gpt2".into(),
        external_id: "ext_1".into(),
        owner: "diptanu@tensorland.ai".into(),
        namespace: "langtech".into(),
        ml_framework: 1,
        created_at: now(),
        updated_at: now(),
    };

    repository
        .create_exeperiment(experiment.clone())
        .await
        .unwrap();

    let maybe_experiment_out = repository.get_experiment("abcd".into()).await.unwrap();
    assert!(maybe_experiment_out.is_some());
    let experiment_out = maybe_experiment_out.unwrap();
    assert_eq!(experiment_out, experiment);
}
