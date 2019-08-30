import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { ItemImage } from '.';

const stories = storiesOf('components/ItemImage', module);
const mockUrl = 'https://i.gyazo.com/8560fce19556b64c95ad091350910184.jpg';

stories
  .add('default', () => (
    <ItemImage imageUrl={mockUrl} title="title" isSoldOut={false} width={300} />
  ))
  .add('sold out', () => (
    <ItemImage
      imageUrl={mockUrl}
      title="sold out"
      isSoldOut={true}
      width={300}
    />
  ))
  .add('image not found', () => (
    <ItemImage
      imageUrl={'fake'}
      title="sold out"
      isSoldOut={true}
      width={300}
    />
  ))
  .add('default(500px)', () => (
    <ItemImage imageUrl={mockUrl} title="title" isSoldOut={false} width={500} />
  ))
  .add('sold out(500px)', () => (
    <ItemImage imageUrl={mockUrl} title="title" isSoldOut={true} width={500} />
  ));
