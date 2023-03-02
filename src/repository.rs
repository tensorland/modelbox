use sea_orm::ActiveValue::NotSet;
use sea_orm::{
    ActiveModelTrait, ColumnTrait, Database, DatabaseConnection, DatabaseTransaction, DbErr,
    EntityTrait, QueryFilter, Set, TransactionTrait,
};

use sea_query::OnConflict;
use time::{OffsetDateTime, PrimitiveDateTime};

use entity::events;
use entity::events::Entity as EventEntity;
use entity::experiments;
use entity::experiments::Entity as ExperimentEntity;
use entity::files;
use entity::files::Entity as FileEntity;
use entity::metadata;
use entity::metadata::Entity as MetadataEntity;
use entity::metrics;
use entity::metrics::Entity as MetricEntity;
use entity::model_versions;
use entity::model_versions::Entity as ModelVersionEntity;
use entity::models;
use entity::models::Entity as ModelEntity;
use entity::mutations;
use thiserror::Error;

pub struct CreateExperimentResult {
    pub exists: bool,
    pub id: String,
}

pub struct CreateModelResult {
    pub exists: bool,
    pub id: String,
}

#[derive(Debug, Default)]
pub struct CreateModelVersionResult {
    pub exists: bool,
    pub id: String,
}

#[derive(Error, Debug)]
pub enum DatastoreError {
    #[error("data store error `{0}`")]
    DatabaseError(#[from] DbErr),

    #[error("serde json error `{0}`")]
    JsonError(#[from] serde_json::Error),
}

#[allow(dead_code)]
enum MutationObject {
    Unknown,
    Experiment,
    Model,
    ModelVersion,
}

#[allow(dead_code)]
enum MutationType {
    Unknown,
    Create,
    Modify,
    Update,
    Delete,
}

#[derive(Debug)]
pub struct Repository {
    conn: DatabaseConnection,
}

fn now() -> PrimitiveDateTime {
    let n = OffsetDateTime::now_utc();
    PrimitiveDateTime::new(n.date(), n.time())
}

impl Repository {
    pub async fn new(db_url: &str) -> Result<Self, DbErr> {
        let db = Database::connect(db_url).await?;
        Ok(Self { conn: db })
    }

    // This is for integration tests 
    #[allow(dead_code)]
    pub fn new_with_db(db: DatabaseConnection) -> Self {
        Self { conn: db }
    }

    pub async fn create_exeperiment(
        &self,
        experiment: experiments::Model,
    ) -> Result<CreateExperimentResult, DatastoreError> {
        let tx = self.conn.begin().await?;
        Self::create_experiment_mutation_event(&tx, experiment.clone()).await?;
        let experiment_id = experiment.id.clone();
        let experiment = experiments::ActiveModel {
            id: Set(experiment.id),
            name: Set(experiment.name),
            external_id: Set(experiment.external_id),
            owner: Set(experiment.owner),
            namespace: Set(experiment.namespace),
            ml_framework: Set(experiment.ml_framework),
            created_at: Set(experiment.created_at),
            updated_at: Set(experiment.updated_at),
        };
        let result = ExperimentEntity::insert(experiment)
            .on_conflict(
                OnConflict::column(experiments::Column::Id)
                    .do_nothing()
                    .to_owned(),
            )
            .exec(&tx)
            .await;
        if let Err(DbErr::RecordNotInserted) = result {
            tx.rollback().await?;
            return Ok(CreateExperimentResult {
                exists: true,
                id: experiment_id,
            });
        }
        tx.commit().await?;
        Ok(CreateExperimentResult {
            exists: false,
            id: experiment_id,
        })
    }

    pub async fn get_experiment(&self, id: &str) -> Result<Option<experiments::Model>, DbErr> {
        ExperimentEntity::find_by_id(id).one(&self.conn).await
    }

    pub async fn list_experiments(
        &self,
        namespace: String,
    ) -> Result<Vec<experiments::Model>, DbErr> {
        let experiment_models = ExperimentEntity::find()
            .filter(experiments::Column::Namespace.eq(namespace))
            .all(&self.conn)
            .await?;
        Ok(experiment_models)
    }

    pub async fn create_model(
        &self,
        model: models::Model,
    ) -> Result<CreateModelResult, DatastoreError> {
        let tx = self.conn.begin().await?;
        Self::create_model_mutation_event(&tx, model.clone()).await?;
        let model_id = model.id.clone();
        let model = models::ActiveModel {
            id: Set(model.id),
            name: Set(model.name),
            owner: Set(model.owner),
            namespace: Set(model.namespace),
            task: Set(model.task),
            description: Set(model.description),
            created_at: Set(model.created_at),
            updated_at: Set(model.updated_at),
        };
        let result = ModelEntity::insert(model)
            .on_conflict(
                OnConflict::column(models::Column::Id)
                    .do_nothing()
                    .to_owned(),
            )
            .exec(&tx)
            .await;
        if let Err(DbErr::RecordNotInserted) = result {
            tx.rollback().await?;
            return Ok(CreateModelResult {
                exists: true,
                id: model_id,
            });
        }
        tx.commit().await?;
        Ok(CreateModelResult {
            exists: false,
            id: model_id,
        })
    }

    pub async fn models_by_namespace(
        &self,
        namespace: String,
    ) -> Result<Vec<models::Model>, DbErr> {
        let models = ModelEntity::find()
            .filter(models::Column::Namespace.eq(namespace))
            .all(&self.conn)
            .await?;
        Ok(models)
    }

    pub async fn create_model_version(
        &self,
        model_version: model_versions::Model,
    ) -> Result<CreateModelVersionResult, DatastoreError> {
        let tx = self.conn.begin().await?;
        Self::create_model_version_mutation_event(&tx, model_version.clone()).await?;
        let mut result = CreateModelVersionResult{ exists: false, id: model_version.id.clone()};
        let model_version = model_versions::ActiveModel {
            id: Set(model_version.id),
            name: Set(model_version.name),
            model_id: Set(model_version.model_id),
            experiment_id: Set(model_version.experiment_id),
            namespace: Set(model_version.namespace),
            version: Set(model_version.version),
            description: Set(model_version.description),
            ml_framework: Set(model_version.ml_framework),
            unique_tags: Set(model_version.unique_tags),
            created_at: Set(model_version.created_at),
            updated_at: Set(model_version.updated_at),
        };
        let query_res= ModelVersionEntity::insert(model_version)
            .on_conflict(
                OnConflict::column(model_versions::Column::Id)
                    .do_nothing()
                    .to_owned(),
            )
            .exec(&tx)
            .await;
        if let Err(DbErr::RecordNotInserted) = query_res {
            tx.rollback().await?;
            result.exists = true;
            return Ok(result);
        }
        tx.commit().await?;
        Ok(result)
    }

    pub async fn model_versions_for_model(
        &self,
        model_id: String,
    ) -> Result<Vec<model_versions::Model>, DbErr> {
        let model_versions = ModelVersionEntity::find()
            .filter(model_versions::Column::ModelId.eq(model_id))
            .all(&self.conn)
            .await?;
        Ok(model_versions)
    }

    pub async fn update_metadata(&self, meta: Vec<metadata::Model>) -> Result<(), DbErr> {
        let mut meta_list: Vec<metadata::ActiveModel> = Vec::new();
        for m in meta {
            let meta = metadata::ActiveModel {
                id: Set(m.id),
                parent_id: Set(m.parent_id),
                name: Set(m.name),
                meta: Set(m.meta),
                created_at: Set(m.created_at),
                updated_at: Set(m.updated_at),
            };
            meta_list.push(meta);
        }
        MetadataEntity::insert_many(meta_list)
            .on_conflict(
                sea_query::OnConflict::column(metadata::Column::Id)
                    .update_column(metadata::Column::Meta)
                    .to_owned(),
            )
            .exec(&self.conn)
            .await?;
        Ok(())
    }

    pub async fn get_metadata(&self, parent_id: String) -> Result<Vec<metadata::Model>, DbErr> {
        let metadata = MetadataEntity::find()
            .filter(metadata::Column::ParentId.eq(parent_id))
            .all(&self.conn)
            .await?;
        Ok(metadata)
    }

    pub async fn create_files(&self, files: Vec<files::Model>) -> Result<(), DbErr> {
        let mut files_list: Vec<files::ActiveModel> = Vec::new();
        for f in files {
            let file = files::ActiveModel {
                id: Set(f.id),
                parent_id: Set(f.parent_id),
                src_path: Set(f.src_path),
                upload_path: Set(f.upload_path),
                metadata: Set(f.metadata),
                file_type: Set(f.file_type),
                artifact_name: Set(f.artifact_name),
                artifact_id: Set(f.artifact_id),
                created_at: Set(f.created_at),
                updated_at: Set(f.updated_at),
            };
            files_list.push(file);
        }
        let result = files::Entity::insert_many(files_list)
            .on_conflict(
                OnConflict::column(files::Column::Id)
                    .update_column(files::Column::UploadPath)
                    .update_column(files::Column::Metadata)
                    .update_column(files::Column::UpdatedAt)
                    .to_owned(),
            )
            .exec(&self.conn)
            .await;

        if let Err(DbErr::RecordNotInserted) = result {
            return Ok(());
        }
        Ok(())
    }

    pub async fn get_files(&self, parent_id: String) -> Result<Vec<files::Model>, DbErr> {
        let files = FileEntity::find()
            .filter(files::Column::ParentId.eq(parent_id))
            .all(&self.conn)
            .await?;
        Ok(files)
    }

    pub async fn create_events(&self, events: Vec<events::Model>) -> Result<(), DbErr> {
        let mut events_list: Vec<events::ActiveModel> = Vec::new();
        for e in events {
            let event = events::ActiveModel {
                id: Set(e.id),
                parent_id: Set(e.parent_id),
                name: Set(e.name),
                source: Set(e.source),
                metadata: Set(e.metadata),
                source_wall_clock: Set(e.source_wall_clock),
            };
            events_list.push(event);
        }
        EventEntity::insert_many(events_list)
            .exec(&self.conn)
            .await?;
        Ok(())
    }

    pub async fn events_for_object(&self, parent_id: String) -> Result<Vec<events::Model>, DbErr> {
        let events = EventEntity::find()
            .filter(events::Column::ParentId.eq(parent_id))
            .all(&self.conn)
            .await?;
        Ok(events)
    }

    pub async fn log_metrics(&self, metrics: Vec<metrics::Model>) -> Result<(), DbErr> {
        let mut metrics_list: Vec<metrics::ActiveModel> = Vec::new();
        for m in metrics {
            let metric = metrics::ActiveModel {
                id: NotSet,
                object_id: Set(m.object_id),
                name: Set(m.name),
                tensor: Set(m.tensor),
                double_value: Set(m.double_value),
                step: Set(m.step),
                wall_clock: Set(m.wall_clock),
                created_at: Set(m.created_at),
            };
            metrics_list.push(metric);
        }
        MetricEntity::insert_many(metrics_list)
            .exec(&self.conn)
            .await?;
        Ok(())
    }

    pub async fn metrics(&self, object_id: String) -> Result<Vec<metrics::Model>, DbErr> {
        let metrics = MetricEntity::find()
            .filter(metrics::Column::ObjectId.eq(object_id))
            .all(&self.conn)
            .await?;
        Ok(metrics)
    }

    pub async fn create_experiment_mutation_event(
        tx: &DatabaseTransaction,
        experiment: experiments::Model,
    ) -> Result<(), DatastoreError> {
        let json_payload = serde_json::to_value(&experiment)?;
        let mutation_event = mutations::ActiveModel {
            id: NotSet,
            object_id: Set(experiment.id),
            object_type: Set(MutationObject::Experiment as i16),
            mutation_type: Set(MutationType::Create as i16),
            namespace: Set(experiment.namespace),
            experiment_payload: Set(Some(json_payload)),
            model_payload: NotSet,
            model_version_payload: NotSet,
            created_at: Set(now()),
            processed_at: Set(None),
        };
        mutation_event.save(tx).await?;
        Ok(())
    }

    pub async fn create_model_mutation_event(
        tx: &DatabaseTransaction,
        model: models::Model,
    ) -> Result<(), DatastoreError> {
        let json_payload = serde_json::to_value(&model)?;
        let mutation_event = mutations::ActiveModel {
            id: NotSet,
            object_id: Set(model.id),
            object_type: Set(MutationObject::Model as i16),
            mutation_type: Set(MutationType::Create as i16),
            namespace: Set(model.namespace),
            experiment_payload: NotSet,
            model_payload: Set(Some(json_payload)),
            model_version_payload: NotSet,
            created_at: Set(now()),
            processed_at: Set(None),
        };
        mutation_event.save(tx).await?;
        Ok(())
    }

    pub async fn create_model_version_mutation_event(
        tx: &DatabaseTransaction,
        model_version: model_versions::Model,
    ) -> Result<(), DatastoreError> {
        let json_payload = serde_json::to_value(&model_version)?;
        let mutation_event = mutations::ActiveModel {
            id: NotSet,
            object_id: Set(model_version.id),
            object_type: Set(MutationObject::ModelVersion as i16),
            mutation_type: Set(MutationType::Create as i16),
            namespace: Set(model_version.namespace),
            experiment_payload: NotSet,
            model_payload: NotSet,
            model_version_payload: Set(Some(json_payload)),
            created_at: Set(now()),
            processed_at: Set(None),
        };
        mutation_event.save(tx).await?;
        Ok(())
    }
}
