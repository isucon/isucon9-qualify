import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { Item } from '.';
import { MemoryRouter } from 'react-router-dom';
import { ItemStatus } from '../../dataObjects/item';
import { GridList } from '@material-ui/core';

const stories = storiesOf('components/Item', module);

const mockProps = {
  itemId: 1,
  imageUrl: 'https://i.gyazo.com/8560fce19556b64c95ad091350910184.jpg',
  title: 'サンプル',
  price: 10000,
  status: 'on_sale' as ItemStatus,
};

stories
  .add('default', () => (
    <MemoryRouter>
      <GridList>
        <Item {...mockProps} />
      </GridList>
    </MemoryRouter>
  ))
  .add('sold out', () => (
    <MemoryRouter>
      <GridList>
        <Item {...mockProps} status="sold_out" />
      </GridList>
    </MemoryRouter>
  ));
