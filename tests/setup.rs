use entity::experiments::Entity as ExperimentEntity;
use entity::mutations::Entity as MutationEntity;
use sea_orm::entity::prelude::*;
use sea_orm::{Database, DbBackend, DbErr, Schema};
use sea_query::table::TableCreateStatement;

pub async fn create_db() -> Result<DatabaseConnection, DbErr> {
    let db = Database::connect("sqlite::memory:").await?;

    setup_schema(&db).await?;

    Ok(db)
}

async fn setup_schema(db: &DbConn) -> Result<(), DbErr> {
    // Setup Schema helper
    let schema = Schema::new(DbBackend::Sqlite);

    // Derive from Entity
    let stmt1: TableCreateStatement = schema.create_table_from_entity(ExperimentEntity);
    let stmt2: TableCreateStatement = schema.create_table_from_entity(MutationEntity);

    // Execute create table statement
    db.execute(db.get_database_backend().build(&stmt1)).await?;
    db.execute(db.get_database_backend().build(&stmt2)).await?;
    Ok(())
}
