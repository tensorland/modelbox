use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_table(
                Table::create()
                    .table(Experiment::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Experiment::Id)
                            .string_len(40)
                            .not_null()
                            .primary_key(),
                    )
                    .col(ColumnDef::new(Experiment::Name).string().not_null())
                    .col(ColumnDef::new(Experiment::ExternalId).string().not_null())
                    .col(ColumnDef::new(Experiment::Owner).string().not_null())
                    .col(ColumnDef::new(Experiment::Namespace).string().not_null())
                    .col(ColumnDef::new(Experiment::MLFramework).integer().not_null())
                    .col(ColumnDef::new(Experiment::CreatedAt).date_time().not_null())
                    .col(ColumnDef::new(Experiment::UpdatedAt).date_time().not_null())
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Model::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Model::Id)
                            .string_len(40)
                            .not_null()
                            .primary_key(),
                    )
                    .col(ColumnDef::new(Model::Name).string().not_null())
                    .col(ColumnDef::new(Model::Owner).string().not_null())
                    .col(ColumnDef::new(Model::Namespace).string().not_null())
                    .col(ColumnDef::new(Model::Task).string().not_null())
                    .col(ColumnDef::new(Model::Description).string().not_null())
                    .col(ColumnDef::new(Model::CreatedAt).date_time().not_null())
                    .col(ColumnDef::new(Model::UpdatedAt).date_time().not_null())
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(ModelVersion::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(ModelVersion::Id)
                            .string_len(40)
                            .not_null()
                            .primary_key(),
                    )
                    .col(ColumnDef::new(ModelVersion::Name).string().not_null())
                    .col(
                        ColumnDef::new(ModelVersion::ModelId)
                            .string_len(40)
                            .not_null(),
                    )
                    .col(ColumnDef::new(ModelVersion::Namespace).string().not_null())
                    .col(
                        ColumnDef::new(ModelVersion::ExperimentId)
                            .string_len(40)
                            .not_null(),
                    )
                    .col(ColumnDef::new(ModelVersion::Version).string().not_null())
                    .col(
                        ColumnDef::new(ModelVersion::Description)
                            .string()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(ModelVersion::MLFramework)
                            .integer()
                            .not_null(),
                    )
                    .col(ColumnDef::new(ModelVersion::UniqueTags).json().not_null())
                    .col(
                        ColumnDef::new(ModelVersion::CreatedAt)
                            .date_time()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(ModelVersion::UpdatedAt)
                            .date_time()
                            .not_null(),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Metadata::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Metadata::Id)
                            .string_len(40)
                            .not_null()
                            .primary_key(),
                    )
                    .col(ColumnDef::new(Metadata::ParentId).string_len(40).not_null())
                    .col(ColumnDef::new(Metadata::Name).string().not_null())
                    .col(ColumnDef::new(Metadata::Meta).json().not_null())
                    .col(ColumnDef::new(Metadata::CreatedAt).date_time().not_null())
                    .col(ColumnDef::new(Metadata::UpdatedAt).date_time().not_null())
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Files::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Files::Id)
                            .string_len(40)
                            .not_null()
                            .primary_key(),
                    )
                    .col(ColumnDef::new(Files::ParentId).string_len(40).not_null())
                    .col(ColumnDef::new(Files::SrcPath).string().not_null())
                    .col(ColumnDef::new(Files::UploadPath).string())
                    .col(ColumnDef::new(Files::FileType).string().not_null())
                    .col(ColumnDef::new(Files::Metadata).json().not_null())
                    .col(ColumnDef::new(Files::ArtifactName).string().not_null())
                    .col(ColumnDef::new(Files::ArtifactId).string_len(40).not_null())
                    .col(ColumnDef::new(Files::CreatedAt).date_time().not_null())
                    .col(ColumnDef::new(Files::UpdatedAt).date_time().not_null())
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Events::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Events::Id)
                            .string_len(40)
                            .not_null()
                            .primary_key(),
                    )
                    .col(ColumnDef::new(Events::ParentId).string_len(40).not_null())
                    .col(ColumnDef::new(Events::Name).string().not_null())
                    .col(ColumnDef::new(Events::Source).string().not_null())
                    .col(ColumnDef::new(Events::Metadata).json().not_null())
                    .col(
                        ColumnDef::new(Events::SourceWallClock)
                            .date_time()
                            .not_null(),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Mutations::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Mutations::Id)
                            .integer()
                            .auto_increment()
                            .not_null()
                            .primary_key(),
                    )
                    .col(
                        ColumnDef::new(Mutations::ObjectId)
                            .string_len(40)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Mutations::ObjectType)
                            .small_unsigned()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Mutations::MutationType)
                            .small_unsigned()
                            .not_null(),
                    )
                    .col(ColumnDef::new(Mutations::Namespace).string().not_null())
                    .col(ColumnDef::new(Mutations::ExperimentPayload).json())
                    .col(ColumnDef::new(Mutations::ModelPayload).json())
                    .col(ColumnDef::new(Mutations::ModelVersionPayload).json())
                    .col(ColumnDef::new(Mutations::CreatedAt).date_time().not_null())
                    .col(ColumnDef::new(Mutations::ProcessedAt).date_time())
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Metrics::Table)
                    .if_not_exists()
                    .col(ColumnDef::new(Metrics::Id).integer().auto_increment().not_null().primary_key())
                    .col(ColumnDef::new(Metrics::ObjectId).string_len(40).not_null())
                    .col(ColumnDef::new(Metrics::Name).string().not_null())
                    .col(ColumnDef::new(Metrics::Tensor).string())
                    .col(ColumnDef::new(Metrics::DoubleValue).double())
                    .col(ColumnDef::new(Metrics::Step).big_unsigned())
                    .col(ColumnDef::new(Metrics::WallClock).date_time())
                    .col(ColumnDef::new(Metrics::CreatedAt).date_time().not_null())
                    .to_owned(),
            )
            .await
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(Table::drop().table(Experiment::Table).to_owned())
            .await
    }
}

#[derive(Iden)]
enum Experiment {
    #[iden = "experiments"]
    Table,
    Id,
    Name,
    ExternalId,
    Owner,
    Namespace,
    MLFramework,
    CreatedAt,
    UpdatedAt,
}

#[derive(Iden)]
enum Model {
    #[iden = "models"]
    Table,
    Id,
    Name,
    Owner,
    Namespace,
    Task,
    Description,
    CreatedAt,
    UpdatedAt,
}

#[derive(Iden)]
enum ModelVersion {
    #[iden = "model_versions"]
    Table,
    Id,
    Name,
    ModelId,
    ExperimentId,
    Namespace,
    Version,
    Description,
    MLFramework,
    UniqueTags,
    CreatedAt,
    UpdatedAt,
}

#[derive(Iden)]
enum Metadata {
    #[iden = "metadata"]
    Table,
    Id,
    ParentId,
    Name,
    Meta,
    CreatedAt,
    UpdatedAt,
}

#[derive(Iden)]
enum Files {
    #[iden = "files"]
    Table,
    Id,
    ParentId,
    SrcPath,
    UploadPath,
    FileType,
    Metadata,
    ArtifactName,
    ArtifactId,
    CreatedAt,
    UpdatedAt,
}

#[derive(Iden)]
enum Events {
    #[iden = "events"]
    Table,
    Id,
    ParentId,
    Name,
    Source,
    Metadata,
    SourceWallClock,
}

#[derive(Iden)]
enum Mutations {
    #[iden = "mutations"]
    Table,
    Id,
    ObjectId,
    ObjectType,
    MutationType,
    Namespace,
    ExperimentPayload,
    ModelPayload,
    ModelVersionPayload,
    CreatedAt,
    ProcessedAt,
}

#[derive(Iden)]
enum Metrics {
    #[iden = "metrics"]
    Table,
    Id,
    WallClock,
    ObjectId,
    Name,
    DoubleValue,
    Tensor,
    Step,
    CreatedAt,
}
