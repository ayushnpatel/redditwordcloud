'use server';

import axios from 'axios';

interface GetRedditWordsFromThreadAPIResponse {
  Words: String[];
  Link: string;
}

export async function makeRequest(
  link: string
): Promise<GetRedditWordsFromThreadAPIResponse> {
  // ...
  try {
    console.log('da linkies', process.env['REDDITWORDCLOUD_API_URL']);
    const response = await axios.post(
      // `${process.env['REDDITWORDCLOUD_API_URL']}\/reddit\/words\/link` ?? '',
      `http://0.0.0.0:8080\/reddit\/words\/link` ?? '',
      {
        link: link,
      },
      {
        headers: { 'Content-Type': 'application/json' },
        data: { link: link },
      }
    );
    console.log(response.status);
    console.log(response.statusText);
    console.log(response.headers);
  } catch (error) {
    // Handle error
    console.error(error);
  }

  return {
    Words: [],
    Link: 'placeholder',
  };
}
