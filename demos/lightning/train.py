import os
import logging
import torch
from torch import nn
import torch.nn.functional as F
import torch.utils.data as data
from torchvision import transforms
from torchvision.datasets import MNIST
from torch.utils.data import DataLoader, random_split
import pytorch_lightning as pl

from model_box_logger import ModelBoxLogger
from pytorch_lightning.callbacks import ModelCheckpoint


logging.getLogger("pytorch_lightning").setLevel(logging.INFO)

class Encoder(nn.Module):
    def __init__(self):
        super().__init__()
        self.l1 = nn.Sequential(nn.Linear(28 * 28, 64), nn.ReLU(), nn.Linear(64, 3))

    def forward(self, x):
        return self.l1(x)


class Decoder(nn.Module):
    def __init__(self):
        super().__init__()
        self.l1 = nn.Sequential(nn.Linear(3, 64), nn.ReLU(), nn.Linear(64, 28 * 28))

    def forward(self, x):
        return self.l1(x)

class LitAutoEncoder(pl.LightningModule):
    def __init__(self, encoder, decoder):
        super().__init__()
        self.encoder = encoder
        self.decoder = decoder

    def training_step(self, batch, batch_idx):
        # training_step defines the train loop.
        x, y = batch
        x = x.view(x.size(0), -1)
        z = self.encoder(x)
        x_hat = self.decoder(z)
        loss = F.mse_loss(x_hat, x)
        return loss

    def test_step(self, batch, batch_idx):
        # training_step defines the train loop.
        x, y = batch
        x = x.view(x.size(0), -1)
        z = self.encoder(x)
        x_hat = self.decoder(z)
        loss = F.mse_loss(x_hat, x)
        self.log("test_loss", loss)

    def validation_step(self, batch, batch_idx):
        # training_step defines the train loop.
        x, y = batch
        x = x.view(x.size(0), -1)
        z = self.encoder(x)
        x_hat = self.decoder(z)
        loss = F.mse_loss(x_hat, x)
        self.log("val_loss", loss)

    def configure_optimizers(self):
        optimizer = torch.optim.Adam(self.parameters(), lr=1e-3)
        return optimizer

train_set = MNIST(os.getcwd(), download=True, transform=transforms.ToTensor(),
        train=True)

# use 20% of training data for validation
train_set_size = int(len(train_set) * 0.8)
valid_set_size = len(train_set) - train_set_size

# split the train set into two
seed = torch.Generator().manual_seed(42)
train_set, valid_set = data.random_split(train_set, [train_set_size, valid_set_size], generator=seed)

test_dataset = MNIST(os.getcwd(), download=True, transform=transforms.ToTensor(),
        train=False)

train_loader = DataLoader(train_set)
val_loader = DataLoader(valid_set)

# model
autoencoder = LitAutoEncoder(Encoder(), Decoder())

# train model
if __name__ == "__main__":
    mbox_logger = ModelBoxLogger("langtech", "lid_quartznet", "diptanuc@gmail.com")

    # train model
    checkpoint_callback = ModelCheckpoint(dirpath="/tmp/checkpoints", save_top_k=100, monitor="val_loss")
    trainer = pl.Trainer(accelerator="gpu", devices=1,enable_checkpointing=True, logger=[mbox_logger],callbacks=[checkpoint_callback])
    trainer.fit(autoencoder, train_loader, val_loader)

    trainer.test(autoencoder, dataloaders=DataLoader(test_dataset))

