import { useState } from 'react';
import { searchItem, Item } from '~/api';

interface Prop {
  onSearchCompleted: (results: Item[]) => void;
}

export const Search = ({ onSearchCompleted }: Prop) => {
  const [keyword, setKeyword] = useState('');

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setKeyword(event.target.value);
  };

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (keyword.trim() === '') {
        onSearchCompleted([]); // キーワードが空なら全件表示に戻す
        return;
      }

      const data = await searchItem(keyword);
      onSearchCompleted(data.items);
    } catch (error) {
      console.error('Search error:', error);
      alert('Failed to search for items');
    }
  };

  return (
    <div className="Search">
      <form onSubmit={onSubmit}>
        <div>
          <button type="submit" className="button">search</button>
          <input
            type="text"
            name="keyword"
            id="keyword"
            placeholder="keyword"
            value={keyword}
            onChange={onValueChange}
          />
        </div>
      </form>
    </div>
  );
};
