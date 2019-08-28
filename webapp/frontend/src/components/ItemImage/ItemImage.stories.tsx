import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { ItemImage } from '.';

const stories = storiesOf('components/ItemImage', module);
const mockUrl = 'https://i.gyazo.com/8560fce19556b64c95ad091350910184.jpg';

stories
  .add('default', () => (
    <ItemImage imageUrl={mockUrl} title="title" isSoldOut={false} />
  ))
  .add('sold out', () => (
    <ItemImage imageUrl={mockUrl} title="sold out" isSoldOut={true} />
  ))
  .add('image not found', () => (
    <ItemImage imageUrl={'fake'} title="sold out" isSoldOut={true} />
  ));
